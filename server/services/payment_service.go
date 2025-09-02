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
	InitiatePayment(id int) (domain.PaymentResponse, error)
	HandleWebhook(webhookData map[string]any) error
}

type PaymentService struct {
	repo                  repositories.PaymentRepositoryInterface
	vendingMachineService VendingMachineServiceInterface
}

func NewPaymentService(repo repositories.PaymentRepositoryInterface, vendingMachineService VendingMachineServiceInterface) *PaymentService {
	return &PaymentService{
		repo:                  repo,
		vendingMachineService: vendingMachineService,
	}
}

func (s *PaymentService) InitiatePayment(id int) (domain.PaymentResponse, error) {
	item, err := s.repo.FetchVendingMachineItem(id)
	if err != nil {
		log.Printf("Failed to fetch vending machine item: %v", err)
		return domain.PaymentResponse{}, err
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
		return domain.PaymentResponse{}, err
	}

	if err = s.repo.CreateTransaction(req.TxRef, item.ID, item.Price); err != nil {
		log.Printf("Failed to create transaction: %v", err)
		return domain.PaymentResponse{}, err
	}

	return domain.PaymentResponse{
		CheckoutURL: checkoutURL,
		TextRef:     req.TxRef,
	}, nil
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
	var combinationID int = 1 // This should come from the webhook data or be retrieved from pending transaction

	// Update transaction status and decrease item amount
	err := s.repo.UpdateStatus(txRef, combinationID)
	if err != nil {
		log.Printf("Failed to update transaction status: %v", err)
		return err
	}

	// Extract vending machine ID from metadata and start the motor
	if vendingMachineIDStr, ok := meta["vending_machine_id"].(string); ok {
		// Convert string to int (you might want to add proper error handling here)
		// For now, assuming it's a valid integer string
		if vendingMachineID, err := strconv.ParseInt(vendingMachineIDStr, 10, 64); err == nil {
			// Start the motor to dispense the item
			if err := s.vendingMachineService.StartMotor(int(vendingMachineID)); err != nil {
				log.Printf("Failed to start motor for vending machine %d: %v", vendingMachineID, err)
				// Don't return error here as the payment was successful
				// Just log the issue for monitoring
			} else {
				log.Printf("Successfully started motor for vending machine %d", vendingMachineID)
			}
		} else {
			log.Printf("Invalid vending machine ID format: %s", vendingMachineIDStr)
		}
	}

	log.Printf("Successfully processed webhook for tx_ref: %s", txRef)
	return nil
}
