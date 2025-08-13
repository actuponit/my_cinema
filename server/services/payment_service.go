package services

import (
	"cinema-server/domain"
	"cinema-server/repositories"
	"errors"
	"log"
	"strconv"

	"github.com/google/uuid"
)

type PaymentServiceInterface interface {
	InitiatePayment(req int) (string, error)
	HandleWebhook(webhookData map[string]any) error
}

type PaymentService struct {
	repo repositories.PaymentRepositoryInterface
}

func NewPaymentService(repo repositories.PaymentRepositoryInterface) *PaymentService {
	return &PaymentService{
		repo: repo,
	}
}

func (s *PaymentService) InitiatePayment(id int) (string, error) {
	item, err := s.repo.FetchVendingMachineItem(id)
	if err != nil {
		log.Printf("Failed to fetch vending machine item: %v", err)
		return "", err
	}
	req := domain.PaymentRequest{
		Amount:           strconv.FormatFloat(item.Price, 'f', -1, 64),
		TxRef:            uuid.New().String(),
		VendingMachineID: id,
	}

	// Call repository to initiate payment with Chapa
	checkoutURL, err := s.repo.InitiatePayment(req)
	if err != nil {
		log.Printf("Failed to initiate payment: %v", err)
		return "", err
	}

	if err = s.repo.CreateTransaction(req.TxRef, item.ID, item.Price); err != nil {
		log.Printf("Failed to create transaction: %v", err)
		return "", err
	}

	return checkoutURL, nil
}

func (s *PaymentService) HandleWebhook(webhookData map[string]any) error {
	// Extract required data from webhook
	txRef, ok := webhookData["tx_ref"].(string)
	if !ok {
		return errors.New("transaction reference not found in webhook data")
	}

	status, ok := webhookData["status"].(string)
	if !ok {
		return errors.New("status not found in webhook data")
	}

	// Check if payment was successful
	if status != "success" {
		log.Printf("Payment failed for tx_ref: %s, status: %s", txRef, status)
		return errors.New("payment was not successful")
	}

	// Extract vending machine and item information from metadata
	meta, ok := webhookData["meta"].(map[string]any)
	if !ok {
		return errors.New("metadata not found in webhook data")
	}

	// Validate vending machine ID exists in metadata
	_, ok = meta["vending_machine_id"].(string)
	if !ok {
		return errors.New("vending machine ID not found in metadata")
	}

	// For now, we'll need to get the combination ID from somewhere
	// This might need to be stored in the webhook metadata or retrieved from a pending transaction
	// For demonstration, assuming we have this value
	combinationID := 1 // This should come from the webhook data or be retrieved from pending transaction

	// Update transaction status and decrease item amount
	err := s.repo.UpdateStatus(txRef, combinationID)
	if err != nil {
		log.Printf("Failed to update transaction status: %v", err)
		return err
	}

	log.Printf("Successfully processed webhook for tx_ref: %s", txRef)
	return nil
}
