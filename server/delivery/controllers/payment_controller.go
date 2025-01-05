package controllers

import (
	"bytes"
	"cinema-server/config"
	"cinema-server/domain"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentController struct{}

func NewPaymentController() *PaymentController {
	return &PaymentController{}
}

func (pc *PaymentController) InitiatePayment(c *gin.Context) {
	// Logic to initiate payment using Chapa

	var req domain.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chapaConfig := config.NewChapaConfig()

	// TODO: Register this uuid in the backend as pending transaction after I get the checout url
	req.TxRef = uuid.New().String()

	payload := map[string]interface{}{
		"amount":                     req.Amount,
		"currency":                   req.Currency,
		"tx_ref":                     req.TxRef,
		"customization[title]":       "Payment for my favourite merchant",
		"customization[description]": "I love online payments",
	}

	client := &http.Client{}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
		return
	}

	r, err := http.NewRequest("POST", chapaConfig.BaseURL+"transaction/initialize", bytes.NewBuffer(jsonPayload))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+chapaConfig.SecretKey)

	resp, err := client.Do(r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "failed to initiate payment"})
		return
	}

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment initiated", "data": respBody})
}

func (pc *PaymentController) ChapaWebhook(c *gin.Context) {
	// verify the request is from chapa

	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request body"})
		return
	}

	// change the status of the request in the database

	// send a firebase notification to the arduino (push notfication)

	fmt.Println(string(reqBodyJSON))

	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}
