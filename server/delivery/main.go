package main

import (
	"cinema-server/delivery/controllers"
	routers "cinema-server/delivery/routes"
	"cinema-server/repositories"
	"cinema-server/services"
	"cinema-server/utils"
	"log"
	"net/http"
	"os"

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

	userService := services.NewUserService(userRepository)
	paymentService := services.NewPaymentService(paymentRepository)

	authController := controllers.NewAuthController(userService)
	fileController := controllers.NewFileController()
	paymentController := controllers.NewPaymentController(paymentService)

	routers.CreateAuthRoutes(router, authController)
	routers.CreateFileRoutes(router, fileController)
	routers.CreatePaymentRoutes(router, paymentController)

	if err := router.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}

	log.Println("Server started on port:", os.Getenv("PORT"))
}
