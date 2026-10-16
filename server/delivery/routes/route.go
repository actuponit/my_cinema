package routers

import (
	"cinema-server/config"
	"cinema-server/delivery/controllers"
	"cinema-server/utils"

	"github.com/gin-gonic/gin"
)

func CreateAuthRoutes(router *gin.Engine, controller *controllers.AuthController) {
	router.POST("auth/login", controller.Login)
	router.POST("auth/register", controller.Register)
}

func CreatePaymentRoutes(router *gin.Engine, controller *controllers.PaymentController) {
	router.POST("payments/initialize", controller.InitiatePayment)
	router.POST("payments/webhook", controller.ChapaWebhook)
	router.POST("payment/initiate", controller.TestGraphqlAction)
}

func CreateWalletRoutes(router *gin.Engine, controller *controllers.WalletController, cfg *config.WalletConfig) {
	wallet := router.Group("v1/wallet")
	auth := utils.SupabaseAuth(cfg)
	wallet.POST("onboard", auth, controller.Onboard)
	wallet.POST("topups", auth, controller.StartTopup)
	wallet.GET("topups/:id", auth, controller.GetTopup)
	wallet.PUT("machine-credentials", auth, controller.SetMachineCredentials)
	wallet.POST("pay", utils.ActionSecret(cfg.ActionSecret), controller.Pay)
	wallet.POST("webhooks/verify", controller.VerifyWebhook)
}
