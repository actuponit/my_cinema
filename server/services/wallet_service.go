package services

import (
	"cinema-server/config"
	"cinema-server/domain"
	"cinema-server/repositories"
	"cinema-server/utils"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type WalletServiceInterface interface {
	Onboard(userID string) (*domain.Wallet, *string, error)
	StartTopup(userID string, requestedAmount *string) (domain.Topup, error)
	GetTopup(userID, topupID string) (domain.Topup, *domain.Wallet, error)
	HandleVerifyWebhook(depositID string) error
	ReconcileOpenTopups() error
	SetMachineCredentials(userID, username, password string) error
	Pay(machineID, username, password string, items []domain.ItemRequest) domain.PayResult
}

const (
	failureLimit  = 30
	failureWindow = time.Minute
	maxTopupCents = 1_000_000 // 10,000.00 ETB
)

var amountPattern = regexp.MustCompile(`^(\d{1,7})(?:\.(\d{1,2}))?$`)

type WalletService struct {
	repo     repositories.WalletRepositoryInterface
	verify   repositories.VerifyCheckoutClientInterface
	vending  VendingMachineServiceInterface
	cfg      *config.WalletConfig
	now      func() time.Time
	compare  func(hash, password []byte) error
	dummyPwd []byte

	// ponytail: in-memory, per process; move to Redis/DB if the backend runs >1 instance
	mu       sync.Mutex
	failures []time.Time
}

func NewWalletService(repo repositories.WalletRepositoryInterface, verify repositories.VerifyCheckoutClientInterface,
	vending VendingMachineServiceInterface, cfg *config.WalletConfig, now func() time.Time) *WalletService {
	dummy, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("generate dummy hash: %v", err))
	}
	return &WalletService{
		repo: repo, verify: verify, vending: vending, cfg: cfg, now: now,
		compare: bcrypt.CompareHashAndPassword, dummyPwd: dummy,
	}
}

func (s *WalletService) Onboard(userID string) (*domain.Wallet, *string, error) {
	w, err := s.repo.Onboard(userID)
	if err != nil {
		return nil, nil, err
	}
	if w == nil {
		return nil, nil, domain.ErrNotCustomer
	}
	username, err := s.repo.GetCredentialUsername(userID)
	if err != nil {
		return nil, nil, err
	}
	return w, username, nil
}

func (s *WalletService) StartTopup(userID string, requestedAmount *string) (domain.Topup, error) {
	w, err := s.repo.GetWallet(userID)
	if err != nil {
		return domain.Topup{}, err
	}
	if w == nil {
		return domain.Topup{}, domain.ErrNotCustomer
	}
	if requestedAmount != nil && !validTopupAmount(*requestedAmount) {
		return domain.Topup{}, domain.ErrInvalidInput
	}

	topupID := uuid.NewString()
	d, err := s.verify.CreateDeposit(topupID, userID, s.cfg.WalletWebURL+"/topup/return?ref="+topupID)
	if err != nil {
		return domain.Topup{}, fmt.Errorf("create deposit: %w", err)
	}
	t := domain.Topup{
		ID: topupID, UserID: userID, DepositID: d.ID, Status: d.Status,
		RequestedAmount: requestedAmount, CheckoutURL: &d.CheckoutURL, ExpiresAt: d.ExpiresAt,
	}
	if err := s.repo.InsertTopup(t); err != nil {
		// A later webhook or reconcile adopts the deposit through wallet_apply_deposit.
		log.Printf("ERROR insert topup failed deposit=%s user=%s: %v", d.ID, userID, err)
		return domain.Topup{}, err
	}
	return t, nil
}

func (s *WalletService) GetTopup(userID, topupID string) (domain.Topup, *domain.Wallet, error) {
	if _, err := uuid.Parse(topupID); err != nil {
		return domain.Topup{}, nil, domain.ErrNotFound
	}
	t, err := s.repo.GetTopup(topupID)
	if err != nil {
		return domain.Topup{}, nil, err
	}
	if t == nil || t.UserID != userID {
		return domain.Topup{}, nil, domain.ErrNotFound
	}
	if !domain.IsTerminalDepositStatus(t.Status) {
		if fresh, err := s.applyFromProvider(t.DepositID); err != nil {
			log.Printf("topup %s refresh failed: %v", t.ID, err)
		} else {
			t = fresh
		}
	}
	w, err := s.repo.GetWallet(userID)
	if err != nil {
		return domain.Topup{}, nil, err
	}
	return *t, w, nil
}

func (s *WalletService) HandleVerifyWebhook(depositID string) error {
	_, err := s.applyFromProvider(depositID)
	return err
}

func (s *WalletService) ReconcileOpenTopups() error {
	now := s.now()
	open, err := s.repo.ListOpenTopups(now.Add(-2*time.Minute), now.Add(-48*time.Hour), 50)
	if err != nil {
		return err
	}
	for _, t := range open {
		if _, err := s.applyFromProvider(t.DepositID); err != nil {
			log.Printf("reconcile topup %s (deposit %s): %v", t.ID, t.DepositID, err)
		}
	}
	return nil
}

// applyFromProvider re-fetches the deposit; status and amount are never taken from a webhook body.
func (s *WalletService) applyFromProvider(depositID string) (*domain.Topup, error) {
	d, err := s.verify.GetDeposit(depositID)
	if err != nil {
		return nil, err
	}
	var amount *string
	if d.Status == "succeeded" {
		amount = &d.Amount
	}
	return s.repo.ApplyDeposit(d.ID, d.CustomerID, d.Status, amount)
}

func (s *WalletService) SetMachineCredentials(userID, username, password string) error {
	w, err := s.repo.GetWallet(userID)
	if err != nil {
		return err
	}
	if w == nil {
		return domain.ErrNotCustomer
	}
	if !domain.ValidCredentialPart(username) || !domain.ValidCredentialPart(password) {
		return domain.ErrInvalidInput
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.repo.UpsertCredential(userID, username, string(hash))
}

func (s *WalletService) Pay(machineID, username, password string, items []domain.ItemRequest) domain.PayResult {
	payItems, ok := sumPayItems(items)
	if _, err := uuid.Parse(machineID); err != nil || !ok {
		return payResult(domain.PayInvalidItems, "These items are not available on this machine.")
	}
	if !domain.ValidCredentialPart(username) || !domain.ValidCredentialPart(password) {
		return payResult(domain.PayInvalidCredentials, "Wrong username or password.")
	}
	if s.tooManyFailures() {
		return payResult(domain.PayTryLater, "Too many attempts right now. Please wait a minute.")
	}

	cred, err := s.repo.GetCredentialByUsername(username)
	if err != nil {
		log.Printf("ERROR wallet pay: get credential: %v", err)
		return payResult(domain.PayError, "Something went wrong. You were not charged.")
	}
	if cred == nil {
		_ = s.compare(s.dummyPwd, []byte(password)) // equal timing for unknown usernames
		s.recordFailure()
		return payResult(domain.PayInvalidCredentials, "Wrong username or password.")
	}
	if cred.LockedUntil != nil && cred.LockedUntil.After(s.now()) {
		minutes := int(cred.LockedUntil.Sub(s.now()).Minutes()) + 1
		return payResult(domain.PayLocked, fmt.Sprintf("Too many wrong attempts. Try again in %d minutes.", minutes))
	}
	if s.compare([]byte(cred.PasswordHash), []byte(password)) != nil {
		if err := s.repo.RecordLoginFailure(cred.UserID); err != nil {
			log.Printf("ERROR wallet pay: record login failure user=%s: %v", cred.UserID, err)
		}
		s.recordFailure()
		return payResult(domain.PayInvalidCredentials, "Wrong username or password.")
	}

	if !s.vending.IsConnected() {
		return payResult(domain.PayError, "Machine is offline. You were not charged.")
	}

	texRef := uuid.NewString()
	res, err := s.repo.Pay(cred.UserID, machineID, payItems, texRef)
	if err != nil {
		log.Printf("ERROR wallet pay tex_ref=%s user=%s: %v", texRef, cred.UserID, err)
		return payResult(domain.PayError, "Something went wrong. You were not charged.")
	}
	switch res.Status {
	case domain.PaySuccess:
	case domain.PayInsufficientBalance:
		res.Message = fmt.Sprintf("Not enough balance. You have %s ETB, this costs %s ETB. Top up in the AutoSuq Wallet app.", res.Balance, res.Total)
		return res
	case domain.PayOutOfStock:
		return payResult(res.Status, "Some items just sold out. Please update your cart.")
	case domain.PayInvalidItems:
		return payResult(res.Status, "These items are not available on this machine.")
	default: // NO_WALLET can't happen: a credential requires a wallet.
		log.Printf("ERROR wallet pay tex_ref=%s user=%s: unexpected status %s", texRef, cred.UserID, res.Status)
		return payResult(domain.PayError, "Something went wrong. You were not charged.")
	}

	if err := s.vending.SendCommand(res.MotorCodes); err != nil {
		// ponytail: no automatic refund (D9); add wallet_refund(tex_ref) if dispense failures show up in logs
		log.Printf("ERROR wallet dispense failed tex_ref=%s user=%s: %v", texRef, cred.UserID, err)
	}
	res.MotorCodes = ""
	res.Message = "Payment successful. Dispensing your items."
	return res
}

func payResult(status, message string) domain.PayResult {
	return domain.PayResult{Status: status, Message: message}
}

// sumPayItems validates kiosk items (1–20 lines, slot id > 0, 1–10 each) and merges duplicate slots.
func sumPayItems(items []domain.ItemRequest) ([]domain.PayItem, bool) {
	if len(items) == 0 || len(items) > 20 {
		return nil, false
	}
	var out []domain.PayItem
	index := map[int]int{}
	for _, it := range items {
		if it.ItemID <= 0 || it.Amount < 1 || it.Amount > 10 {
			return nil, false
		}
		if i, seen := index[it.ItemID]; seen {
			out[i].Quantity += it.Amount
			continue
		}
		index[it.ItemID] = len(out)
		out = append(out, domain.PayItem{SlotID: it.ItemID, Quantity: it.Amount})
	}
	return out, true
}

func validTopupAmount(s string) bool {
	m := amountPattern.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	whole, _ := strconv.ParseInt(m[1], 10, 64)
	frac := m[2]
	for len(frac) < 2 {
		frac += "0"
	}
	cents, _ := strconv.ParseInt(frac, 10, 64)
	total := whole*100 + cents
	return total > 0 && total <= maxTopupCents
}

func (s *WalletService) prune(now time.Time) {
	i := 0
	for i < len(s.failures) && now.Sub(s.failures[i]) >= failureWindow {
		i++
	}
	s.failures = s.failures[i:]
}

func (s *WalletService) tooManyFailures() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prune(s.now())
	return len(s.failures) >= failureLimit
}

func (s *WalletService) recordFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.prune(now)
	s.failures = append(s.failures, now)
}
