package main

import (
	"cinema-server/config"
	"cinema-server/delivery/controllers"
	routers "cinema-server/delivery/routes"
	"cinema-server/repositories"
	"cinema-server/services"
	"cinema-server/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hasura/go-graphql-client"
)

type hasuraAuthTransport struct {
	AdminSecret string
	Transport   http.RoundTripper
}

// RoundTrip adds the x-hasura-admin-secret header to every request
func (h *hasuraAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-hasura-admin-secret", h.AdminSecret)
	return h.Transport.RoundTrip(req)
}

func main() {
	walletConfig, err := config.NewWalletConfig()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.New()
	router.Use(utils.CORSMiddleware())
	router.Use(gin.Logger())

	adminSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")
	hasuraEndpoint := os.Getenv("HASURA_GRAPHQL_ENDPOINT")
	httpClient := &http.Client{
		Transport: &hasuraAuthTransport{
			AdminSecret: adminSecret,
			Transport:   http.DefaultTransport,
		},
	}

	client := graphql.NewClient(hasuraEndpoint, httpClient)

	userRepository := repositories.NewUserRepository(client)
	paymentRepository := repositories.NewPaymentRepository(client)
	walletRepository := repositories.NewWalletRepository(client)
	verifyClient := repositories.NewVerifyCheckoutClient(walletConfig.VerifyAPIBaseURL, walletConfig.VerifyAPIKey)

	// One persistent connection; if the broker is down, paho keeps retrying in the background.
	mqttClient := utils.NewMQTTClientFromConfig(*config.NewMQTTConfig())
	go func() {
		if err := mqttClient.Connect(); err != nil {
			log.Printf("MQTT connect failed: %v", err)
		}
	}()

	// Initialize services
	userService := services.NewUserService(userRepository)
	vendingMachineService := services.NewVendingMachineService(mqttClient)
	paymentService := services.NewPaymentService(paymentRepository, vendingMachineService)
	walletService := services.NewWalletService(walletRepository, verifyClient, vendingMachineService, walletConfig, time.Now)

	go func() {
		for range time.NewTicker(60 * time.Second).C {
			if err := walletService.ReconcileOpenTopups(); err != nil {
				log.Printf("reconcile open topups: %v", err)
			}
		}
	}()

	authController := controllers.NewAuthController(userService)
	paymentController := controllers.NewPaymentController(paymentService)
	walletController := controllers.NewWalletController(walletService, walletConfig.VerifyWebhookSecret)

	routers.CreateAuthRoutes(router, authController)
	routers.CreatePaymentRoutes(router, paymentController)
	routers.CreateWalletRoutes(router, walletController, walletConfig)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Println("Server starting on port:", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
