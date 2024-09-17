package main

import (
	"cinema-server/delivery/controllers"
	routers "cinema-server/delivery/routes"
	"cinema-server/domain"
	"cinema-server/repositories"
	"cinema-server/services"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	router := gin.New()
	router.Use(gin.Logger())

	userRepository := repositories.NewUserRepository([]domain.User{})
	userService := services.NewUserService(userRepository)

	authController := controllers.NewAuthController(userService)

	routers.CreateAuthRoutes(router, authController)

	if err := router.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}

	log.Println("Server started on port:", os.Getenv("PORT"))

}
