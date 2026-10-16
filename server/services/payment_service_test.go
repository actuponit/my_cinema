package services

import (
	"cinema-server/domain"
	"cinema-server/repositories"
	"testing"
)

type fakePaymentRepo struct {
	repositories.PaymentRepositoryInterface // unused methods panic
	verification                            domain.ChapaVerification
	basket                                  []domain.BasketLine
	paid                                    bool
}

func (f *fakePaymentRepo) VerifyChapaTransaction(string) (domain.ChapaVerification, error) {
	return f.verification, nil
}
func (f *fakePaymentRepo) FetchBasket(string) ([]domain.BasketLine, error) { return f.basket, nil }

// MarkChapaPaid mimics chapa_mark_paid: rows only on the first false -> true flip.
func (f *fakePaymentRepo) MarkChapaPaid(string) ([]domain.VendingMachineItem, error) {
	if f.paid {
		return nil, nil
	}
	f.paid = true
	return []domain.VendingMachineItem{{ID: 1, Amount: 2, MotorCode: "A1"}, {ID: 2, Amount: 1, MotorCode: "B3"}}, nil
}

func intPtr(i int) *int { return &i }

func newChapaFixture() (*PaymentService, *fakePaymentRepo, *fakeVending) {
	repo := &fakePaymentRepo{
		verification: domain.ChapaVerification{Status: "success", Amount: "45.5", Currency: "ETB"},
		basket:       []domain.BasketLine{{Price: "10.00", Amount: intPtr(2)}, {Price: "25.50", Amount: intPtr(1)}},
	}
	vending := &fakeVending{connected: true}
	return NewPaymentService(repo, vending), repo, vending
}

func TestChapaWebhookReplayDispensesOnce(t *testing.T) {
	svc, _, vending := newChapaFixture()
	for i := 0; i < 2; i++ {
		if err := svc.HandleWebhook("tx_1"); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	if len(vending.sent) != 1 || vending.sent[0] != "A1,A1,B3" {
		t.Fatalf("SendCommand calls = %v", vending.sent)
	}
}

func TestChapaWebhookRejectsUnpaidOrMismatched(t *testing.T) {
	for name, mutate := range map[string]func(*fakePaymentRepo){
		"failed status":   func(r *fakePaymentRepo) { r.verification.Status = "failed" },
		"wrong currency":  func(r *fakePaymentRepo) { r.verification.Currency = "USD" },
		"amount mismatch": func(r *fakePaymentRepo) { r.verification.Amount = "45.49" },
		"empty basket":    func(r *fakePaymentRepo) { r.basket = nil },
	} {
		t.Run(name, func(t *testing.T) {
			svc, repo, vending := newChapaFixture()
			mutate(repo)
			if err := svc.HandleWebhook("tx_1"); err != nil {
				t.Fatal(err)
			}
			if repo.paid || len(vending.sent) != 0 {
				t.Fatalf("marked paid=%v sent=%v", repo.paid, vending.sent)
			}
		})
	}
}
