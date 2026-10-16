package services

import (
	"cinema-server/domain"
	"cinema-server/repositories"
	"errors"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type PaymentServiceInterface interface {
	InitiatePayment(items []domain.ItemRequest) (domain.PaymentResponse, error)
	HandleWebhook(txRef string) error
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
	}

	txRef := uuid.New().String()
	req := domain.PaymentRequest{
		Amount:           strconv.FormatFloat(totalPrice, 'f', 2, 64),
		TxRef:            txRef,
		VendingMachineID: itemIDs,
		Currency:         "ETB",
	}

	// Call repository to initiate payment with Chapa
	checkoutURL, err := s.repo.InitiatePayment(req)
	if err != nil {
		log.Printf("Failed to initiate payment: %v", err)
		return domain.PaymentResponse{}, err
	}

	// Create transactions for all items with their specific amounts
	if err = s.repo.CreateTransactions(txRef, vendingItems, requestedAmounts); err != nil {
		log.Printf("Failed to create transactions: %v", err)
		return domain.PaymentResponse{}, err
	}

	return domain.PaymentResponse{
		CheckoutURL: checkoutURL,
		TextRef:     txRef,
	}, nil
}

// HandleWebhook re-verifies the payment with Chapa and dispenses the basket exactly once.
// It returns nil for payments that must not be retried (unpaid, mismatched, already dispensed).
func (s *PaymentService) HandleWebhook(txRef string) error {
	if txRef == "" {
		return errors.New("transaction reference not found in webhook data")
	}

	v, err := s.repo.VerifyChapaTransaction(txRef)
	if err != nil {
		return err
	}
	if v.Status != "success" || v.Currency != "ETB" {
		log.Printf("chapa tx_ref=%s not paid: status=%s currency=%s", txRef, v.Status, v.Currency)
		return nil
	}

	basket, err := s.repo.FetchBasket(txRef)
	if err != nil {
		return err
	}
	expected, ok := basketTotal(basket)
	paid, okPaid := new(big.Rat).SetString(v.Amount)
	if !ok || !okPaid || len(basket) == 0 || expected.Cmp(paid) != 0 {
		log.Printf("ERROR chapa tx_ref=%s amount mismatch: paid=%s basket=%d rows", txRef, v.Amount, len(basket))
		return nil
	}

	items, err := s.repo.MarkChapaPaid(txRef)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		log.Printf("chapa tx_ref=%s already processed", txRef)
		return nil
	}

	var codes []string
	for _, item := range items {
		for j := 0; j < item.Amount && item.MotorCode != ""; j++ {
			codes = append(codes, item.MotorCode)
		}
	}
	if len(codes) == 0 {
		log.Printf("ERROR chapa tx_ref=%s: paid but no motor codes to dispense", txRef)
		return nil
	}
	if err := s.vendingMachineService.SendCommand(strings.Join(codes, ",")); err != nil {
		// Paid and marked; don't fail the webhook (a retry would not dispense again anyway).
		log.Printf("ERROR chapa dispense failed tx_ref=%s: %v", txRef, err)
	}
	return nil
}

// basketTotal is sum(price * coalesce(amount, 1)), in exact decimals.
func basketTotal(lines []domain.BasketLine) (*big.Rat, bool) {
	total := new(big.Rat)
	for _, l := range lines {
		price, ok := new(big.Rat).SetString(l.Price)
		if !ok {
			return nil, false
		}
		qty := int64(1)
		if l.Amount != nil {
			qty = int64(*l.Amount)
		}
		total.Add(total, price.Mul(price, big.NewRat(qty, 1)))
	}
	return total, true
}
