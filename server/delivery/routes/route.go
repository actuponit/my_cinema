package routers

import (
	"cinema-server/delivery/controllers"

	"github.com/gin-gonic/gin"
)

func CreateAuthRoutes(router *gin.Engine, controller *controllers.AuthController) {
	router.POST("auth/login", controller.Login)
	router.POST("auth/register", controller.Register)
}
