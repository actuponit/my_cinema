package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
}

func (pc *PaymentController) InitiatePayment(c *gin.Context) {
	// Logic to initiate payment using Chapa

	c.JSON(http.StatusOK, gin.H{"message": "Payment initiated"})
}

func (pc *PaymentController) ChapaWebhook(c *gin.Context) {
	// Logic to handle Chapa webhook
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}

func RegisterPaymentRoutes(router *gin.Engine) {
	pc := &PaymentController{}
	router.POST("/initiate-payment", pc.InitiatePayment)
	router.POST("/chapa-webhook", pc.ChapaWebhook)
}
