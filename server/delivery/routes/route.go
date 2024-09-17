package routers

import (
	"cinema-server/delivery/controllers"
	"cinema-server/utils"

	"github.com/gin-gonic/gin"
)

func CreateAuthRoutes(router *gin.Engine, controller *controllers.AuthController) {
	router.POST("auth/login", utils.AuthMiddleware("user"), controller.Login)
	router.POST("auth/register", controller.Register)
}
