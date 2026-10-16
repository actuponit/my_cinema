package repositories

import (
	"cinema-server/domain"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hasura/go-graphql-client"
)

type WalletRepositoryInterface interface {
	Onboard(userID string) (*domain.Wallet, error)
	GetWallet(userID string) (*domain.Wallet, error)
	InsertTopup(t domain.Topup) error
	GetTopup(id string) (*domain.Topup, error)
	ListOpenTopups(olderThan, newerThan time.Time, limit int) ([]domain.Topup, error)
	ApplyDeposit(depositID, userID, status string, amount *string) (*domain.Topup, error)
	GetCredentialByUsername(username string) (*domain.MachineCredential, error)
	GetCredentialUsername(userID string) (*string, error)
	UpsertCredential(userID, username, passwordHash string) error
	RecordLoginFailure(userID string) error
	Pay(userID, machineID string, items []domain.PayItem, texRef string) (domain.PayResult, error)
}

// WalletRepository talks to Hasura with the admin secret. Numerics arrive as strings
// (HASURA_GRAPHQL_STRINGIFY_NUMERIC_TYPES=true) and stay strings.
type WalletRepository struct {
	client *graphql.Client
}

func NewWalletRepository(client *graphql.Client) *WalletRepository {
	return &WalletRepository{client: client}
}

func (r *WalletRepository) exec(query string, vars map[string]any, out any) error {
	data, err := r.client.ExecRaw(context.Background(), query, vars)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode hasura response: %w", err)
	}
	return nil
}

type walletRow struct {
	UserID  string `json:"user_id"`
	Balance string `json:"balance"`
}

func (w walletRow) toDomain() *domain.Wallet {
	return &domain.Wallet{UserID: w.UserID, Balance: w.Balance}
}

const topupFields = `id user_id deposit_id status requested_amount credited_amount checkout_url expires_at`

type topupRow struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	DepositID       string     `json:"deposit_id"`
	Status          string     `json:"status"`
	RequestedAmount *string    `json:"requested_amount"`
	CreditedAmount  *string    `json:"credited_amount"`
	CheckoutURL     *string    `json:"checkout_url"`
	ExpiresAt       *time.Time `json:"expires_at"`
}

func (t topupRow) toDomain() domain.Topup {
	return domain.Topup{
		ID: t.ID, UserID: t.UserID, DepositID: t.DepositID, Status: t.Status,
		RequestedAmount: t.RequestedAmount, CreditedAmount: t.CreditedAmount,
		CheckoutURL: t.CheckoutURL, ExpiresAt: t.ExpiresAt,
	}
}

// Onboard returns nil when the account may not use the wallet (owner/admin).
func (r *WalletRepository) Onboard(userID string) (*domain.Wallet, error) {
	var out struct {
		Rows []walletRow `json:"wallet_onboard"`
	}
	err := r.exec(`mutation($u: uuid!) { wallet_onboard(args: {p_user_id: $u}) { user_id balance } }`,
		map[string]any{"u": userID}, &out)
	if err != nil {
		return nil, fmt.Errorf("wallet_onboard: %w", err)
	}
	if len(out.Rows) == 0 {
		return nil, nil
	}
	return out.Rows[0].toDomain(), nil
}

func (r *WalletRepository) GetWallet(userID string) (*domain.Wallet, error) {
	var out struct {
		Wallet *walletRow `json:"wallets_by_pk"`
	}
	err := r.exec(`query($u: uuid!) { wallets_by_pk(user_id: $u) { user_id balance } }`,
		map[string]any{"u": userID}, &out)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	if out.Wallet == nil {
		return nil, nil
	}
	return out.Wallet.toDomain(), nil
}

func (r *WalletRepository) InsertTopup(t domain.Topup) error {
	var out struct {
		Row *struct {
			ID string `json:"id"`
		} `json:"insert_wallet_topups_one"`
	}
	obj := map[string]any{
		"id": t.ID, "user_id": t.UserID, "deposit_id": t.DepositID, "status": t.Status,
		"requested_amount": t.RequestedAmount, "checkout_url": t.CheckoutURL, "expires_at": t.ExpiresAt,
	}
	err := r.exec(`mutation($o: wallet_topups_insert_input!) { insert_wallet_topups_one(object: $o) { id } }`,
		map[string]any{"o": obj}, &out)
	if err != nil {
		return fmt.Errorf("insert topup: %w", err)
	}
	return nil
}

func (r *WalletRepository) GetTopup(id string) (*domain.Topup, error) {
	var out struct {
		Topup *topupRow `json:"wallet_topups_by_pk"`
	}
	err := r.exec(`query($id: uuid!) { wallet_topups_by_pk(id: $id) { `+topupFields+` } }`,
		map[string]any{"id": id}, &out)
	if err != nil {
		return nil, fmt.Errorf("get topup: %w", err)
	}
	if out.Topup == nil {
		return nil, nil
	}
	t := out.Topup.toDomain()
	return &t, nil
}

func (r *WalletRepository) ListOpenTopups(olderThan, newerThan time.Time, limit int) ([]domain.Topup, error) {
	var out struct {
		Rows []topupRow `json:"wallet_topups"`
	}
	err := r.exec(`query($old: timestamptz!, $new: timestamptz!, $limit: Int!) {
		wallet_topups(
			where: {status: {_nin: ["succeeded", "failed", "expired", "cancelled"]},
			        created_at: {_lt: $old, _gt: $new}},
			order_by: {created_at: asc}, limit: $limit
		) { `+topupFields+` }
	}`, map[string]any{"old": olderThan, "new": newerThan, "limit": limit}, &out)
	if err != nil {
		return nil, fmt.Errorf("list open topups: %w", err)
	}
	topups := make([]domain.Topup, len(out.Rows))
	for i, row := range out.Rows {
		topups[i] = row.toDomain()
	}
	return topups, nil
}

func (r *WalletRepository) ApplyDeposit(depositID, userID, status string, amount *string) (*domain.Topup, error) {
	var out struct {
		Rows []topupRow `json:"wallet_apply_deposit"`
	}
	args := map[string]any{"p_deposit_id": depositID, "p_user_id": userID, "p_status": status, "p_amount": amount}
	err := r.exec(`mutation($a: wallet_apply_deposit_args!) { wallet_apply_deposit(args: $a) { `+topupFields+` } }`,
		map[string]any{"a": args}, &out)
	if err != nil {
		return nil, fmt.Errorf("wallet_apply_deposit %s: %w", depositID, err)
	}
	if len(out.Rows) == 0 {
		return nil, fmt.Errorf("wallet_apply_deposit %s returned no row", depositID)
	}
	t := out.Rows[0].toDomain()
	return &t, nil
}

func (r *WalletRepository) GetCredentialByUsername(username string) (*domain.MachineCredential, error) {
	var out struct {
		Rows []struct {
			UserID         string     `json:"user_id"`
			Username       string     `json:"username"`
			PasswordHash   string     `json:"password_hash"`
			FailedAttempts int        `json:"failed_attempts"`
			LockedUntil    *time.Time `json:"locked_until"`
		} `json:"machine_credentials"`
	}
	err := r.exec(`query($u: String!) {
		machine_credentials(where: {username: {_eq: $u}}, limit: 1) {
			user_id username password_hash failed_attempts locked_until
		}
	}`, map[string]any{"u": username}, &out)
	if err != nil {
		return nil, fmt.Errorf("get credential: %w", err)
	}
	if len(out.Rows) == 0 {
		return nil, nil
	}
	c := out.Rows[0]
	return &domain.MachineCredential{
		UserID: c.UserID, Username: c.Username, PasswordHash: c.PasswordHash,
		FailedAttempts: c.FailedAttempts, LockedUntil: c.LockedUntil,
	}, nil
}

func (r *WalletRepository) GetCredentialUsername(userID string) (*string, error) {
	var out struct {
		Row *struct {
			Username string `json:"username"`
		} `json:"machine_credentials_by_pk"`
	}
	err := r.exec(`query($u: uuid!) { machine_credentials_by_pk(user_id: $u) { username } }`,
		map[string]any{"u": userID}, &out)
	if err != nil {
		return nil, fmt.Errorf("get credential username: %w", err)
	}
	if out.Row == nil {
		return nil, nil
	}
	return &out.Row.Username, nil
}

func (r *WalletRepository) UpsertCredential(userID, username, passwordHash string) error {
	var out struct {
		Row *struct {
			UserID string `json:"user_id"`
		} `json:"insert_machine_credentials_one"`
	}
	obj := map[string]any{
		"user_id": userID, "username": username, "password_hash": passwordHash,
		"failed_attempts": 0, "locked_until": nil, "updated_at": time.Now().UTC(),
	}
	err := r.exec(`mutation($o: machine_credentials_insert_input!) {
		insert_machine_credentials_one(object: $o, on_conflict: {
			constraint: machine_credentials_pkey,
			update_columns: [username, password_hash, failed_attempts, locked_until, updated_at]
		}) { user_id }
	}`, map[string]any{"o": obj}, &out)
	if err != nil {
		if strings.Contains(err.Error(), "machine_credentials_username_key") {
			return domain.ErrUsernameTaken
		}
		return fmt.Errorf("upsert credential: %w", err)
	}
	return nil
}

func (r *WalletRepository) RecordLoginFailure(userID string) error {
	var out struct {
		Rows []struct {
			UserID string `json:"user_id"`
		} `json:"machine_login_failed"`
	}
	err := r.exec(`mutation($u: uuid!) { machine_login_failed(args: {p_user_id: $u}) { user_id } }`,
		map[string]any{"u": userID}, &out)
	if err != nil {
		return fmt.Errorf("machine_login_failed: %w", err)
	}
	return nil
}

func (r *WalletRepository) Pay(userID, machineID string, items []domain.PayItem, texRef string) (domain.PayResult, error) {
	var out struct {
		Rows []struct {
			Status     string  `json:"status"`
			TexRef     *string `json:"tex_ref"`
			Total      *string `json:"total"`
			Balance    *string `json:"balance"`
			MotorCodes *string `json:"motor_codes"`
		} `json:"wallet_pay"`
	}
	args := map[string]any{"p_user_id": userID, "p_machine_id": machineID, "p_items": items, "p_tex_ref": texRef}
	err := r.exec(`mutation($a: wallet_pay_args!) { wallet_pay(args: $a) { status tex_ref total balance motor_codes } }`,
		map[string]any{"a": args}, &out)
	if err != nil {
		return domain.PayResult{}, fmt.Errorf("wallet_pay %s: %w", texRef, err)
	}
	if len(out.Rows) != 1 {
		return domain.PayResult{}, fmt.Errorf("wallet_pay %s returned %d rows", texRef, len(out.Rows))
	}
	row := out.Rows[0]
	return domain.PayResult{
		Status: row.Status, TexRef: deref(row.TexRef), Total: deref(row.Total),
		Balance: deref(row.Balance), MotorCodes: deref(row.MotorCodes),
	}, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
