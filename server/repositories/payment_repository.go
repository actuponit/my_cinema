package repositories

import (
	"bytes"
	"cinema-server/config"
	"cinema-server/domain"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hasura/go-graphql-client"
)

type PaymentRepositoryInterface interface {
	FetchVendingMachineItem(id int) (domain.VendingMachineItem, error)
	CreateTransaction(texRef string, id int, price float64) error
	UpdateStatus(texRef string, id int) error
	InitiatePayment(domain.PaymentRequest) (string, error)
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
			ID     int `graphql:"id"`
			Amount int `graphql:"amount"`
			Item   struct {
				Price float64 `graphql:"price"`
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

	return domain.VendingMachineItem{
		ID:     item.ID,
		Amount: item.Amount,
		Price:  item.Item.Price,
	}, nil
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

func (r *PaymentRepository) UpdateStatus(texRef string, id int) error {
	var m struct {
		UpdateTransactions struct {
			AffectedRows int `graphql:"affected_rows"`
		} `graphql:"update_transactions(where: {tex_ref: {_eq: $tex_ref}}, _set: {status: true})"`
		UpdateVendingMachine struct {
			id int `graphql:"id"`
		} `graphql:"update_vending_machine_to_items_by_pk(pk_columns: {id: $id}, _inc: {amount: -1})"`
	}

	variables := map[string]any{
		"tex_ref": texRef,
		"id":      id,
	}

	err := r.client.Mutate(context.Background(), &m, variables)
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
