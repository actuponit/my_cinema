package routers

import (
	"cinema-server/delivery/controllers"
	"cinema-server/utils"

	"github.com/gin-gonic/gin"
)

func CreateAuthRoutes(router *gin.Engine, controller *controllers.AuthController) {
	router.POST("auth/login", controller.Login)
	router.POST("auth/register", controller.Register)
}

func CreateFileRoutes(router *gin.Engine, controller *controllers.FileController) {
	router.POST("uploads", utils.AuthMiddleware("cinema"), utils.FileMiddleware, controller.UploadFiles)
	router.DELETE("uploads", controller.DeleteFile)
}
