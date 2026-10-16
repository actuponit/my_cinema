package repositories

import (
	"bytes"
	"cinema-server/config"
	"cinema-server/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hasura/go-graphql-client"
)

// parsePrice converts a stringified numeric value returned by Hasura
// (HASURA_GRAPHQL_STRINGIFY_NUMERIC_TYPES=true) into a float64.
func parsePrice(s string) (float64, error) {
	price, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price %q: %w", s, err)
	}
	return price, nil
}

type PaymentRepositoryInterface interface {
	FetchVendingMachineItem(id int) (domain.VendingMachineItem, error)
	FetchVendingMachineItems(ids []int) ([]domain.VendingMachineItem, error)
	CreateTransaction(texRef string, id int, price float64) error
	CreateTransactions(texRef string, items []domain.VendingMachineItem, amounts map[int]int) error
	InitiatePayment(domain.PaymentRequest) (string, error)
	VerifyChapaTransaction(txRef string) (domain.ChapaVerification, error)
	FetchBasket(texRef string) ([]domain.BasketLine, error)
	MarkChapaPaid(texRef string) ([]domain.VendingMachineItem, error)
}

type PaymentRepository struct {
	client     *graphql.Client
	httpClient *http.Client
}

func NewPaymentRepository(client *graphql.Client) *PaymentRepository {
	return &PaymentRepository{
		client:     client,
		httpClient: &http.Client{},
	}
}

func (r *PaymentRepository) FetchVendingMachineItem(id int) (domain.VendingMachineItem, error) {
	var q struct {
		VendingMachineToItemsByPk struct {
			ID        int    `graphql:"id"`
			Amount    int    `graphql:"amount"`
			MotorCode string `graphql:"motor_code"`
			Item      struct {
				Price string `graphql:"price"`
			} `graphql:"vending_machine_item"`
		} `graphql:"vending_machine_to_items_by_pk(id: $id)"`
	}

	variables := map[string]any{
		"id": id,
	}

	err := r.client.Query(context.Background(), &q, variables)
	if err != nil {
		return domain.VendingMachineItem{}, err
	}

	// Check if the item exists
	if q.VendingMachineToItemsByPk.ID == 0 {
		return domain.VendingMachineItem{}, errors.New("vending machine item not found")
	}

	item := q.VendingMachineToItemsByPk

	price, err := parsePrice(item.Item.Price)
	if err != nil {
		return domain.VendingMachineItem{}, err
	}

	return domain.VendingMachineItem{
		ID:        item.ID,
		Amount:    item.Amount,
		Price:     price,
		MotorCode: item.MotorCode,
	}, nil
}

func (r *PaymentRepository) FetchVendingMachineItems(ids []int) ([]domain.VendingMachineItem, error) {
	// Use _in operator to fetch multiple items (simpler than _or)
	// This generates: where: {id: {_in: [1, 2, 3]}}
	var q struct {
		VendingMachineToItems []struct {
			ID        int    `graphql:"id"`
			Amount    int    `graphql:"amount"`
			MotorCode string `graphql:"motor_code"`
			Item      struct {
				Price string `graphql:"price"`
			} `graphql:"vending_machine_item"`
		} `graphql:"vending_machine_to_items(where: {id: {_in: $ids}})"`
	}

	variables := map[string]any{
		"ids": ids,
	}

	err := r.client.Query(context.Background(), &q, variables)
	if err != nil {
		return nil, err
	}

	// Check if any items were found
	if len(q.VendingMachineToItems) == 0 {
		return nil, errors.New("no vending machine items found")
	}

	var items []domain.VendingMachineItem
	for _, item := range q.VendingMachineToItems {
		price, err := parsePrice(item.Item.Price)
		if err != nil {
			return nil, err
		}
		items = append(items, domain.VendingMachineItem{
			ID:        item.ID,
			Amount:    item.Amount,
			Price:     price,
			MotorCode: item.MotorCode,
		})
	}

	return items, nil
}

func (r *PaymentRepository) CreateTransaction(texRef string, id int, price float64) error {
	query := `
		mutation InsertTransaction($combination_id: Int!, $price: numeric!, $tex_ref: String!) {
			insert_transactions_one(object: {
				combination_id: $combination_id,
				price: $price,
				tex_ref: $tex_ref
			}) {
				id
			}
		}
	`

	m := struct {
		InsertTransactionsOne struct {
			ID string `graphql:"id"`
		} `graphql:"insert_transactions_one"`
	}{}

	variables := map[string]any{
		"combination_id": id,
		"price":          price,
		"tex_ref":        texRef,
	}

	err := r.client.Exec(context.Background(), query, &m, variables)
	if err != nil {
		return err
	}

	return nil
}

// CreateTransactions stores one row per slot with that slot's unit price.
func (r *PaymentRepository) CreateTransactions(texRef string, items []domain.VendingMachineItem, amounts map[int]int) error {
	// Build the objects array for bulk insert
	type TransactionObject struct {
		CombinationID int     `json:"combination_id"`
		Price         float64 `json:"price"`
		TexRef        string  `json:"tex_ref"`
		Amount        int     `json:"amount"`
	}

	var objects []TransactionObject
	for _, item := range items {
		objects = append(objects, TransactionObject{
			CombinationID: item.ID,
			Price:         item.Price,
			TexRef:        texRef,
			Amount:        amounts[item.ID],
		})
	}

	query := `
		mutation InsertTransactions($objects: [transactions_insert_input!]!) {
			insert_transactions(objects: $objects) {
				affected_rows
			}
		}
	`

	m := struct {
		InsertTransactions struct {
			AffectedRows int `graphql:"affected_rows"`
		} `graphql:"insert_transactions"`
	}{}

	variables := map[string]any{
		"objects": objects,
	}

	err := r.client.Exec(context.Background(), query, &m, variables)
	if err != nil {
		return err
	}

	return nil
}

func (r *PaymentRepository) InitiatePayment(req domain.PaymentRequest) (string, error) {
	chapaConfig := config.NewChapaConfig()

	payload := map[string]interface{}{
		"amount":                     req.Amount,
		"currency":                   req.Currency,
		"tx_ref":                     req.TxRef,
		"customization[title]":       "Payment for vending machine",
		"customization[description]": "Payment for vending machine",
		"meta[vending_machine_id]":   req.VendingMachineID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", chapaConfig.BaseURL+"transaction/initialize", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+chapaConfig.SecretKey)

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to initiate payment")
	}

	var respBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}

	// Return the checkout URL from the response
	if data, ok := respBody["data"].(map[string]any); ok {
		if checkoutURL, ok := data["checkout_url"].(string); ok {
			return checkoutURL, nil
		}
	}

	return "", errors.New("checkout URL not found in response")
}

func (r *PaymentRepository) VerifyChapaTransaction(txRef string) (domain.ChapaVerification, error) {
	chapaConfig := config.NewChapaConfig()
	req, err := http.NewRequest(http.MethodGet, chapaConfig.BaseURL+"transaction/verify/"+url.PathEscape(txRef), nil)
	if err != nil {
		return domain.ChapaVerification{}, err
	}
	req.Header.Set("Authorization", "Bearer "+chapaConfig.SecretKey)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return domain.ChapaVerification{}, fmt.Errorf("chapa verify %s: %w", txRef, err)
	}
	defer resp.Body.Close()

	var body struct {
		Message string `json:"message"`
		Data    *struct {
			Status   string      `json:"status"`
			Amount   json.Number `json:"amount"`
			Currency string      `json:"currency"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return domain.ChapaVerification{}, fmt.Errorf("chapa verify %s: HTTP %d: %w", txRef, resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || body.Data == nil {
		return domain.ChapaVerification{}, fmt.Errorf("chapa verify %s: HTTP %d: %s", txRef, resp.StatusCode, body.Message)
	}
	return domain.ChapaVerification{Status: body.Data.Status, Amount: body.Data.Amount.String(), Currency: body.Data.Currency}, nil
}

func (r *PaymentRepository) FetchBasket(texRef string) ([]domain.BasketLine, error) {
	var out struct {
		Rows []domain.BasketLine `json:"transactions"`
	}
	data, err := r.client.ExecRaw(context.Background(),
		`query($t: String!) { transactions(where: {tex_ref: {_eq: $t}}) { price amount } }`,
		map[string]any{"t": texRef})
	if err != nil {
		return nil, fmt.Errorf("fetch basket %s: %w", texRef, err)
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode basket %s: %w", texRef, err)
	}
	return out.Rows, nil
}

// MarkChapaPaid flips the basket to paid exactly once (chapa_mark_paid). An empty result means
// it was already processed and must not be dispensed again.
func (r *PaymentRepository) MarkChapaPaid(texRef string) ([]domain.VendingMachineItem, error) {
	var out struct {
		Rows []struct {
			CombinationID int  `json:"combination_id"`
			Amount        *int `json:"amount"`
			Slot          struct {
				MotorCode string `json:"motor_code"`
			} `json:"vending_machine_to_item"`
		} `json:"chapa_mark_paid"`
	}
	data, err := r.client.ExecRaw(context.Background(),
		`mutation($t: String!) {
			chapa_mark_paid(args: {p_tex_ref: $t}) { combination_id amount vending_machine_to_item { motor_code } }
		}`, map[string]any{"t": texRef})
	if err != nil {
		return nil, fmt.Errorf("chapa_mark_paid %s: %w", texRef, err)
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode chapa_mark_paid %s: %w", texRef, err)
	}
	items := make([]domain.VendingMachineItem, len(out.Rows))
	for i, row := range out.Rows {
		amount := 1
		if row.Amount != nil {
			amount = *row.Amount
		}
		items[i] = domain.VendingMachineItem{ID: row.CombinationID, Amount: amount, MotorCode: row.Slot.MotorCode}
	}
	return items, nil
}
