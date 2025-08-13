package controllers

import (
	"bytes"
	"cinema-server/config"
	"cinema-server/domain"
	"cinema-server/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentController struct {
	paymentService services.PaymentServiceInterface
}

func NewPaymentController(s services.PaymentServiceInterface) *PaymentController {
	return &PaymentController{paymentService: s}
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
	// Updated ChapaWebhook with MQTT publishing
	var reqBody map[string]any
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request body"})
		return
	}

	// Initialize MQTT client options
	opts := mqtt.NewClientOptions().AddBroker("tcp://test.mosquitto.org:1883")
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})

	// Create and start a client using the above ClientOptions
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to MQTT broker"})
		return
	}
	defer client.Disconnect(250)

	// Define the topic and message
	topic := "chapa/webhook"

	// Publish the message
	if token := client.Publish(topic, 0, false, "start_motor"); token.Wait() && token.Error() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish MQTT message"})
		return
	}

	pc.paymentService.HandleWebhook(reqBody)

	fmt.Println(string(reqBodyJSON))
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received and message published to MQTT"})
}

func (pc *PaymentController) TestGraphqlAction(c *gin.Context) {
	// Define the request structure
	var reqBody struct {
		Id int `json:"id"`
	}

	// Bind the JSON request body
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Print the request
	checkoutURL, err := pc.paymentService.InitiatePayment(reqBody.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the expected JSON response
	c.IndentedJSON(http.StatusOK, gin.H{
		"checkOutUrl": checkoutURL,
	})
}
