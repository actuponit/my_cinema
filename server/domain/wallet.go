package domain

import (
	"errors"
	"regexp"
	"time"
)

// Money is a decimal string ("12.50"); totals are computed in SQL, never in float64.
type Wallet struct {
	UserID  string
	Balance string
}

type Topup struct {
	ID, UserID, DepositID, Status   string
	RequestedAmount, CreditedAmount *string
	CheckoutURL                     *string
	ExpiresAt                       *time.Time
}

type MachineCredential struct {
	UserID, Username, PasswordHash string
	FailedAttempts                 int
	LockedUntil                    *time.Time
}

type PayItem struct {
	SlotID   int `json:"slot_id"`
	Quantity int `json:"quantity"`
}

type PayResult struct {
	Status, TexRef, Total, Balance, MotorCodes string
	Message                                    string // English fallback; the kiosk maps Status to its own copy
}

// Deposit is a verify.et deposit as returned by its API.
type Deposit struct {
	ID, Status, Amount, CustomerID, CheckoutURL string
	ExpiresAt                                   *time.Time
}

const (
	PaySuccess             = "SUCCESS"
	PayInvalidCredentials  = "INVALID_CREDENTIALS"
	PayLocked              = "LOCKED"
	PayInsufficientBalance = "INSUFFICIENT_BALANCE"
	PayOutOfStock          = "OUT_OF_STOCK"
	PayInvalidItems        = "INVALID_ITEMS"
	PayTryLater            = "TRY_LATER"
	PayError               = "ERROR"
)

var (
	ErrNotCustomer   = errors.New("not a wallet customer")
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotFound      = errors.New("not found")
	ErrUsernameTaken = errors.New("username taken")
)

var credentialPattern = regexp.MustCompile(`^[A-Za-z0-9]{4,6}$`)

func ValidCredentialPart(s string) bool {
	return credentialPattern.MatchString(s)
}

func IsTerminalDepositStatus(s string) bool {
	switch s {
	case "succeeded", "failed", "expired", "cancelled":
		return true
	}
	return false
}
