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
	InitiatePayment(items []domain.ItemRequest) (domain.PaymentResponse, error)
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

func (s *PaymentService) InitiatePayment(items []domain.ItemRequest) (domain.PaymentResponse, error) {
	// Extract item IDs from the request
	var itemIDs []int
	requestedAmounts := make(map[int]int) // Map to store requested amounts per item
	for _, item := range items {
		itemIDs = append(itemIDs, item.ItemID)
		requestedAmounts[item.ItemID] = item.Amount
	}

	// Fetch all vending machine items using _or condition
	vendingItems, err := s.repo.FetchVendingMachineItems(itemIDs)
	if err != nil {
		log.Printf("Failed to fetch vending machine items: %v", err)
		return domain.PaymentResponse{}, err
	}

	// Validate that all requested items were found
	if len(vendingItems) != len(itemIDs) {
		log.Printf("Some items were not found in the database")
		return domain.PaymentResponse{}, errors.New("some items not found")
	}

	// Calculate total price based on requested amounts
	var totalPrice float64
	var combinationIDs []int
	for _, vendingItem := range vendingItems {
		requestedAmount, ok := requestedAmounts[vendingItem.ID]
		if !ok {
			log.Printf("Item %d not found in requested amounts", vendingItem.ID)
			return domain.PaymentResponse{}, errors.New("item amount mismatch")
		}

		// Validate that requested amount doesn't exceed available amount
		if requestedAmount > vendingItem.Amount {
			log.Printf("Requested amount %d exceeds available amount %d for item %d",
				requestedAmount, vendingItem.Amount, vendingItem.ID)
			return domain.PaymentResponse{}, errors.New("insufficient item quantity")
		}

		totalPrice += vendingItem.Price * float64(requestedAmount)
		combinationIDs = append(combinationIDs, vendingItem.ID)
	}

	txRef := uuid.New().String()
	req := domain.PaymentRequest{
		Amount:           strconv.FormatFloat(totalPrice, 'f', -1, 64),
		TxRef:            txRef,
		VendingMachineID: combinationIDs,
		Currency:         "ETB",
	}

	// Call repository to initiate payment with Chapa
	checkoutURL, err := s.repo.InitiatePayment(req)
	if err != nil {
		log.Printf("Failed to initiate payment: %v", err)
		return domain.PaymentResponse{}, err
	}

	// Create transactions for all items with their specific amounts
	if err = s.repo.CreateTransactions(txRef, combinationIDs, totalPrice, requestedAmounts); err != nil {
		log.Printf("Failed to create transactions: %v", err)
		return domain.PaymentResponse{}, err
	}

	return domain.PaymentResponse{
		CheckoutURL: checkoutURL,
		TextRef:     txRef,
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

	// Update all transaction statuses and get all items with their motor codes
	items, err := s.repo.UpdateStatusMultiple(txRef)
	if err != nil {
		log.Printf("Failed to update transaction statuses: %v", err)
		return err
	}

	if len(items) == 0 {
		log.Printf("No items found for tx_ref: %s", txRef)
		return errors.New("no items found for the given transaction reference")
	}
	// Build comma-separated list of motor codes from items
	var motorCode string
	for i := range items {
		mc := items[i].MotorCode
		if mc == "" {
			continue
		}
		if motorCode != "" {
			motorCode += ","
		}
		motorCode += mc
	}

	if err := s.vendingMachineService.SendCommand(0, motorCode); err != nil {
		log.Printf("Failed to start motor for vending machine %s: %v", motorCode, err)
		// Don't return error here as the payment was successful
		// Just log the issue for monitoring
	} else {
		log.Printf("Successfully started motor for vending machine %s", motorCode)
	}

	// Extract vending machine ID from metadata and start the motor
	// if vendingMachineIDStr, ok := meta["vending_machine_id"].(string); ok {
	// 	// Convert string to int (you might want to add proper error handling here)
	// 	// For now, assuming it's a valid integer string
	// 	if vendingMachineID, err := strconv.ParseInt(vendingMachineIDStr, 10, 64); err == nil {
	// 		// Start the motor to dispense the item

	// 	} else {
	// 		log.Printf("Invalid vending machine ID format: %s", vendingMachineIDStr)
	// 	}
	// }

	log.Printf("Successfully processed webhook for tx_ref: %s, dispensed %d items", txRef, len(items))
	return nil
}
