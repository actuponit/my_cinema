package services

import (
	"cinema-server/config"
	"cinema-server/domain"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type fakeWalletRepo struct {
	creds        map[string]*domain.MachineCredential
	wallets      map[string]*domain.Wallet
	topups       map[string]*domain.Topup
	payCalls     [][]domain.PayItem
	payResult    domain.PayResult
	failures     int
	applied      []string
	upsertedUser string
	upsertedHash string
}

func (f *fakeWalletRepo) Onboard(string) (*domain.Wallet, error)        { return nil, nil }
func (f *fakeWalletRepo) GetCredentialUsername(string) (*string, error) { return nil, nil }
func (f *fakeWalletRepo) GetWallet(id string) (*domain.Wallet, error)   { return f.wallets[id], nil }
func (f *fakeWalletRepo) InsertTopup(domain.Topup) error                { return nil }
func (f *fakeWalletRepo) GetTopup(id string) (*domain.Topup, error)     { return f.topups[id], nil }
func (f *fakeWalletRepo) ListOpenTopups(time.Time, time.Time, int) ([]domain.Topup, error) {
	return nil, nil
}
func (f *fakeWalletRepo) ApplyDeposit(depositID, userID, status string, amount *string) (*domain.Topup, error) {
	f.applied = append(f.applied, depositID+":"+status)
	return &domain.Topup{ID: "t1", UserID: userID, DepositID: depositID, Status: status, CreditedAmount: amount}, nil
}
func (f *fakeWalletRepo) GetCredentialByUsername(u string) (*domain.MachineCredential, error) {
	return f.creds[u], nil
}
func (f *fakeWalletRepo) UpsertCredential(userID, _, hash string) error {
	f.upsertedUser, f.upsertedHash = userID, hash
	return nil
}
func (f *fakeWalletRepo) RecordLoginFailure(string) error { f.failures++; return nil }
func (f *fakeWalletRepo) Pay(_, _ string, items []domain.PayItem, texRef string) (domain.PayResult, error) {
	f.payCalls = append(f.payCalls, items)
	r := f.payResult
	if r.Status == domain.PaySuccess {
		r.TexRef = texRef
	}
	return r, nil
}

type fakeVerify struct {
	deposit domain.Deposit
	gets    int
}

func (f *fakeVerify) CreateDeposit(string, string, string) (domain.Deposit, error) {
	return f.deposit, nil
}
func (f *fakeVerify) GetDeposit(string) (domain.Deposit, error) {
	f.gets++
	return f.deposit, nil
}

type fakeVending struct {
	connected bool
	sent      []string
}

func (f *fakeVending) IsConnected() bool { return f.connected }
func (f *fakeVending) SendCommand(codes string) error {
	f.sent = append(f.sent, codes)
	return nil
}

const machineID = "7f9c2c4e-2a3b-4f7e-9a10-1b2c3d4e5f60"

type harness struct {
	svc      *WalletService
	repo     *fakeWalletRepo
	verify   *fakeVerify
	vending  *fakeVending
	compares int
	now      time.Time
}

func newHarness(t *testing.T) *harness {
	hash, err := bcrypt.GenerateFromPassword([]byte("Pa55"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{
		repo: &fakeWalletRepo{
			creds:     map[string]*domain.MachineCredential{"Abcd": {UserID: "u1", Username: "Abcd", PasswordHash: string(hash)}},
			wallets:   map[string]*domain.Wallet{"u1": {UserID: "u1", Balance: "50.00"}},
			topups:    map[string]*domain.Topup{},
			payResult: domain.PayResult{Status: domain.PaySuccess, Total: "20.00", Balance: "30.00", MotorCodes: "A1,A1,B2"},
		},
		verify:  &fakeVerify{},
		vending: &fakeVending{connected: true},
		now:     time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC),
	}
	h.svc = NewWalletService(h.repo, h.verify, h.vending, &config.WalletConfig{WalletWebURL: "https://w.example"}, func() time.Time { return h.now })
	h.svc.compare = func(hash, pw []byte) error {
		h.compares++
		return bcrypt.CompareHashAndPassword(hash, pw)
	}
	return h
}

var cart = []domain.ItemRequest{{ItemID: 3, Amount: 1}, {ItemID: 5, Amount: 1}, {ItemID: 3, Amount: 1}}

func TestPaySuccess(t *testing.T) {
	h := newHarness(t)
	res := h.svc.Pay(machineID, "Abcd", "Pa55", cart)
	if res.Status != domain.PaySuccess || res.Total != "20.00" || res.Balance != "30.00" || res.TexRef == "" {
		t.Fatalf("unexpected result %+v", res)
	}
	if len(h.repo.payCalls) != 1 {
		t.Fatalf("Pay called %d times", len(h.repo.payCalls))
	}
	got := h.repo.payCalls[0]
	if len(got) != 2 || got[0] != (domain.PayItem{SlotID: 3, Quantity: 2}) || got[1] != (domain.PayItem{SlotID: 5, Quantity: 1}) {
		t.Fatalf("items not summed: %+v", got)
	}
	if len(h.vending.sent) != 1 || h.vending.sent[0] != "A1,A1,B2" {
		t.Fatalf("SendCommand = %v", h.vending.sent)
	}
}

func TestPayUnknownUsername(t *testing.T) {
	h := newHarness(t)
	res := h.svc.Pay(machineID, "Nobody", "Pa55", cart)
	if res.Status != domain.PayInvalidCredentials || len(h.repo.payCalls) != 0 || h.compares != 1 {
		t.Fatalf("status=%s payCalls=%d compares=%d", res.Status, len(h.repo.payCalls), h.compares)
	}
}

func TestPayWrongPassword(t *testing.T) {
	h := newHarness(t)
	res := h.svc.Pay(machineID, "Abcd", "Wr0ng", cart)
	if res.Status != domain.PayInvalidCredentials || h.repo.failures != 1 || len(h.repo.payCalls) != 0 {
		t.Fatalf("status=%s failures=%d", res.Status, h.repo.failures)
	}
}

func TestPayLocked(t *testing.T) {
	h := newHarness(t)
	until := h.now.Add(10 * time.Minute)
	h.repo.creds["Abcd"].LockedUntil = &until
	res := h.svc.Pay(machineID, "Abcd", "Pa55", cart)
	if res.Status != domain.PayLocked || len(h.repo.payCalls) != 0 || h.compares != 0 {
		t.Fatalf("status=%s payCalls=%d compares=%d", res.Status, len(h.repo.payCalls), h.compares)
	}
}

func TestPayMachineOffline(t *testing.T) {
	h := newHarness(t)
	h.vending.connected = false
	res := h.svc.Pay(machineID, "Abcd", "Pa55", cart)
	if res.Status != domain.PayError || len(h.repo.payCalls) != 0 {
		t.Fatalf("status=%s payCalls=%d", res.Status, len(h.repo.payCalls))
	}
}

func TestPayInsufficientBalance(t *testing.T) {
	h := newHarness(t)
	h.repo.payResult = domain.PayResult{Status: domain.PayInsufficientBalance, Total: "35.50", Balance: "30.00"}
	res := h.svc.Pay(machineID, "Abcd", "Pa55", cart)
	if res.Status != domain.PayInsufficientBalance || res.Total != "35.50" || res.Balance != "30.00" || len(h.vending.sent) != 0 {
		t.Fatalf("res=%+v sent=%v", res, h.vending.sent)
	}
}

func TestPayGlobalLimiter(t *testing.T) {
	h := newHarness(t)
	for i := 0; i < failureLimit; i++ {
		if res := h.svc.Pay(machineID, "Abcd", "Wr0ng", cart); res.Status != domain.PayInvalidCredentials {
			t.Fatalf("attempt %d: %s", i, res.Status)
		}
	}
	if res := h.svc.Pay(machineID, "Abcd", "Pa55", cart); res.Status != domain.PayTryLater {
		t.Fatalf("expected TRY_LATER, got %s", res.Status)
	}
	h.now = h.now.Add(failureWindow)
	if res := h.svc.Pay(machineID, "Abcd", "Pa55", cart); res.Status != domain.PaySuccess {
		t.Fatalf("expected SUCCESS after the window, got %s", res.Status)
	}
}

func TestPayInvalidItems(t *testing.T) {
	h := newHarness(t)
	for _, items := range [][]domain.ItemRequest{nil, {{ItemID: 0, Amount: 1}}, {{ItemID: 1, Amount: 11}}} {
		if res := h.svc.Pay(machineID, "Abcd", "Pa55", items); res.Status != domain.PayInvalidItems {
			t.Fatalf("items %v: %s", items, res.Status)
		}
	}
	if res := h.svc.Pay("not-a-uuid", "Abcd", "Pa55", cart); res.Status != domain.PayInvalidItems {
		t.Fatalf("bad machine id: %s", res.Status)
	}
}

func TestGetTopupRefreshesOnlyWhenOpen(t *testing.T) {
	h := newHarness(t)
	const id = "0e7c7d8a-1111-4222-8333-944455556666"
	h.repo.topups[id] = &domain.Topup{ID: id, UserID: "u1", DepositID: "dep_1", Status: "awaiting_transfer"}
	h.verify.deposit = domain.Deposit{ID: "dep_1", Status: "succeeded", Amount: "75.00", CustomerID: "u1"}

	topup, _, err := h.svc.GetTopup("u1", id)
	if err != nil || h.verify.gets != 1 || topup.Status != "succeeded" || *topup.CreditedAmount != "75.00" {
		t.Fatalf("err=%v gets=%d topup=%+v", err, h.verify.gets, topup)
	}

	h.repo.topups[id].Status = "succeeded"
	if _, _, err := h.svc.GetTopup("u1", id); err != nil || h.verify.gets != 1 {
		t.Fatalf("terminal topup hit the provider: err=%v gets=%d", err, h.verify.gets)
	}

	if _, _, err := h.svc.GetTopup("someone-else", id); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("other user's topup: %v", err)
	}
}

func TestWebhookAmountOnlyWhenSucceeded(t *testing.T) {
	h := newHarness(t)
	h.verify.deposit = domain.Deposit{ID: "dep_2", Status: "review_required", Amount: "40.00", CustomerID: "u1"}
	if err := h.svc.HandleVerifyWebhook("dep_2"); err != nil {
		t.Fatal(err)
	}
	if h.repo.applied[0] != "dep_2:review_required" {
		t.Fatalf("applied %v", h.repo.applied)
	}
}

func TestSetMachineCredentialsValidation(t *testing.T) {
	h := newHarness(t)
	for _, bad := range []string{"abc", "abcdefg", "ab_12"} {
		if err := h.svc.SetMachineCredentials("u1", bad, "Pa55"); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("username %q: %v", bad, err)
		}
		if err := h.svc.SetMachineCredentials("u1", "Abcd", bad); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("password %q: %v", bad, err)
		}
	}
	for _, good := range []string{"Ab12", "A1b2C3"} {
		if err := h.svc.SetMachineCredentials("u1", good, good); err != nil {
			t.Fatalf("%q rejected: %v", good, err)
		}
		if bcrypt.CompareHashAndPassword([]byte(h.repo.upsertedHash), []byte(good)) != nil {
			t.Fatalf("stored hash doesn't match %q", good)
		}
	}
	if err := h.svc.SetMachineCredentials("no-wallet", "Ab12", "Ab12"); !errors.Is(err, domain.ErrNotCustomer) {
		t.Fatalf("no wallet: %v", err)
	}
}

func TestValidTopupAmount(t *testing.T) {
	for s, want := range map[string]bool{
		"50": true, "50.5": true, "10000.00": true, "0.01": true,
		"0": false, "0.00": false, "10000.01": false, "1.234": false, "-5": false, "abc": false, "": false,
	} {
		if got := validTopupAmount(s); got != want {
			t.Errorf("validTopupAmount(%q) = %v, want %v", s, got, want)
		}
	}
}
