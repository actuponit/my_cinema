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
	FetchVendingMachineItems(ids []int) ([]domain.VendingMachineItem, error)
	CreateTransaction(texRef string, id int, price float64) error
	CreateTransactions(texRef string, ids []int, totalPrice float64, amounts map[int]int) error
	UpdateStatus(texRef string, id int) (string, error)
	UpdateStatusMultiple(texRef string) ([]domain.VendingMachineItem, error)
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
			ID        int    `graphql:"id"`
			Amount    int    `graphql:"amount"`
			MotorCode string `graphql:"motor_code"`
			Item      struct {
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
		ID:        item.ID,
		Amount:    item.Amount,
		Price:     item.Item.Price,
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
				Price float64 `graphql:"price"`
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
		items = append(items, domain.VendingMachineItem{
			ID:        item.ID,
			Amount:    item.Amount,
			Price:     item.Item.Price,
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

func (r *PaymentRepository) CreateTransactions(texRef string, ids []int, totalPrice float64, amounts map[int]int) error {
	// Build the objects array for bulk insert
	type TransactionObject struct {
		CombinationID int     `json:"combination_id"`
		Price         float64 `json:"price"`
		TexRef        string  `json:"tex_ref"`
		Amount        int     `json:"amount"`
	}

	var objects []TransactionObject
	for _, id := range ids {
		amount := amounts[id] // Get the specific amount for this item
		objects = append(objects, TransactionObject{
			CombinationID: id,
			Price:         totalPrice,
			TexRef:        texRef,
			Amount:        amount,
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

func (r *PaymentRepository) UpdateStatus(texRef string, id int) (string, error) {
	var m struct {
		UpdateTransactions struct {
			Returning []struct {
				CombinationID        int `graphql:"combination_id"`
				VendingMachineToItem struct {
					MotorCode string `graphql:"motor_code"`
				} `graphql:"vending_machine_to_item"`
			} `graphql:"returning"`
			AffectedRows int `graphql:"affected_rows"`
		} `graphql:"update_transactions(where: {tex_ref: {_eq: $tex_ref}}, _set: {status: true})"`
	}

	variables := map[string]any{
		"tex_ref": texRef,
		// "id":      id,
	}

	err := r.client.Mutate(context.Background(), &m, variables)
	if err != nil {
		return "", err
	}

	return m.UpdateTransactions.Returning[0].VendingMachineToItem.MotorCode, nil
}

func (r *PaymentRepository) UpdateStatusMultiple(texRef string) ([]domain.VendingMachineItem, error) {
	var m struct {
		UpdateTransactions struct {
			Returning []struct {
				CombinationID        int `graphql:"combination_id"`
				VendingMachineToItem struct {
					ID        int    `graphql:"id"`
					Amount    int    `graphql:"amount"`
					MotorCode string `graphql:"motor_code"`
					Item      struct {
						Price float64 `graphql:"price"`
					} `graphql:"vending_machine_item"`
				} `graphql:"vending_machine_to_item"`
			} `graphql:"returning"`
			AffectedRows int `graphql:"affected_rows"`
		} `graphql:"update_transactions(where: {tex_ref: {_eq: $tex_ref}}, _set: {status: true})"`
	}

	variables := map[string]any{
		"tex_ref": texRef,
	}

	err := r.client.Mutate(context.Background(), &m, variables)
	if err != nil {
		return nil, err
	}

	var items []domain.VendingMachineItem
	for _, transaction := range m.UpdateTransactions.Returning {
		items = append(items, domain.VendingMachineItem{
			ID:        transaction.VendingMachineToItem.ID,
			Amount:    transaction.VendingMachineToItem.Amount,
			MotorCode: transaction.VendingMachineToItem.MotorCode,
			Price:     transaction.VendingMachineToItem.Item.Price,
		})
	}

	return items, nil
}

func (r *PaymentRepository) InitiatePayment(req domain.PaymentRequest) (string, error) {
	chapaConfig := config.NewChapaConfig()

	payload := map[string]interface{}{
		"amount":   req.Amount,
		"currency": req.Currency,
		"tx_ref":   req.TxRef,
		// "return_url":                 "vendingapp://vendingapp/payment-callback/" + req.TxRef,
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
