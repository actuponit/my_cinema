package controllers

import (
	"cinema-server/domain"
	"cinema-server/services"
	"cinema-server/utils"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WalletController struct {
	service       services.WalletServiceInterface
	webhookSecret string
}

func NewWalletController(s services.WalletServiceInterface, webhookSecret string) *WalletController {
	return &WalletController{service: s, webhookSecret: webhookSecret}
}

func walletError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"code": code, "message": message})
}

// writeServiceError maps the known service errors; anything else is a 500.
func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotCustomer):
		walletError(c, http.StatusForbidden, "not_customer", "This account can't use the wallet")
	case errors.Is(err, domain.ErrInvalidInput):
		walletError(c, http.StatusBadRequest, "invalid_input", "Invalid input")
	case errors.Is(err, domain.ErrNotFound):
		walletError(c, http.StatusNotFound, "not_found", "Not found")
	case errors.Is(err, domain.ErrUsernameTaken):
		walletError(c, http.StatusConflict, "username_taken", "That username is taken")
	default:
		log.Printf("wallet %s %s: %v", c.Request.Method, c.FullPath(), err)
		walletError(c, http.StatusInternalServerError, "internal_error", "Something went wrong")
	}
}

func (wc *WalletController) Onboard(c *gin.Context) {
	w, username, err := wc.service.Onboard(c.GetString("supabase_user_id"))
	if errors.Is(err, domain.ErrNotCustomer) {
		walletError(c, http.StatusForbidden, "owner_account", "Owner accounts can't use the wallet")
		return
	}
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": w.Balance, "username": username})
}

func (wc *WalletController) StartTopup(c *gin.Context) {
	var req struct {
		RequestedAmount *string `json:"requested_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		walletError(c, http.StatusBadRequest, "invalid_input", "Invalid JSON body")
		return
	}
	t, err := wc.service.StartTopup(c.GetString("supabase_user_id"), req.RequestedAmount)
	if err != nil {
		if errors.Is(err, domain.ErrNotCustomer) || errors.Is(err, domain.ErrInvalidInput) {
			writeServiceError(c, err)
			return
		}
		log.Printf("wallet start topup: %v", err)
		walletError(c, http.StatusBadGateway, "provider_error", "Couldn't start the top-up. Please try again.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"topup_id": t.ID, "checkout_url": t.CheckoutURL, "expires_at": t.ExpiresAt})
}

func (wc *WalletController) GetTopup(c *gin.Context) {
	t, w, err := wc.service.GetTopup(c.GetString("supabase_user_id"), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	var balance *string
	if w != nil {
		balance = &w.Balance
	}
	c.JSON(http.StatusOK, gin.H{
		"topup_id": t.ID, "status": t.Status, "requested_amount": t.RequestedAmount,
		"credited_amount": t.CreditedAmount, "balance": balance,
	})
}

func (wc *WalletController) SetMachineCredentials(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		walletError(c, http.StatusBadRequest, "invalid_input", "Invalid JSON body")
		return
	}
	if err := wc.service.SetMachineCredentials(c.GetString("supabase_user_id"), req.Username, req.Password); err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": req.Username})
}

// Pay is the Hasura payWithWallet action handler; it always answers 200 with PayWithWalletOutput.
func (wc *WalletController) Pay(c *gin.Context) {
	var req struct {
		MachineID string               `json:"machine_id"`
		Username  string               `json:"username"`
		Password  string               `json:"password"`
		Items     []domain.ItemRequest `json:"items"`
	}
	res := domain.PayResult{Status: domain.PayInvalidItems, Message: "These items are not available on this machine."}
	if err := c.ShouldBindJSON(&req); err == nil {
		res = wc.service.Pay(req.MachineID, req.Username, req.Password, req.Items)
	}
	c.JSON(http.StatusOK, gin.H{
		"status": res.Status, "message": res.Message,
		"txRef": nullIfEmpty(res.TexRef), "total": nullIfEmpty(res.Total), "balance": nullIfEmpty(res.Balance),
	})
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

const maxWebhookBody = 256 << 10

func (wc *WalletController) VerifyWebhook(c *gin.Context) {
	raw, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxWebhookBody))
	if err != nil {
		walletError(c, http.StatusRequestEntityTooLarge, "body_too_large", "Body too large")
		return
	}
	err = utils.VerifyCheckoutSignature(raw, c.GetHeader("VerifyCheckout-Timestamp"),
		c.GetHeader("VerifyCheckout-Signature"), wc.webhookSecret, time.Now())
	if err != nil {
		log.Printf("verify.et webhook rejected: %v", err)
		walletError(c, http.StatusUnauthorized, "invalid_signature", "Invalid signature")
		return
	}

	var event struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID string `json:"id"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &event); err != nil {
		log.Printf("verify.et webhook: undecodable body: %v", err)
		c.Status(http.StatusOK)
		return
	}
	if event.Type == "webhook.test" {
		c.Status(http.StatusOK)
		return
	}
	if event.Data.Object.ID == "" {
		log.Printf("verify.et webhook %s (%s): no deposit id", event.ID, event.Type)
		c.Status(http.StatusOK)
		return
	}
	if err := wc.service.HandleVerifyWebhook(event.Data.Object.ID); err != nil {
		log.Printf("verify.et webhook %s deposit %s: %v", event.ID, event.Data.Object.ID, err)
		walletError(c, http.StatusInternalServerError, "processing_error", "Retry later")
		return
	}
	c.Status(http.StatusOK)
}
