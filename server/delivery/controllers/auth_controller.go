package controllers

import (
	"cinema-server/domain"
	"cinema-server/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	userService services.UserServiceInterface
}

func NewAuthController(service services.UserServiceInterface) *AuthController {
	return &AuthController{userService: service}
}

func (c *AuthController) Register(ctx *gin.Context) {
	var newUser domain.User

	if err := ctx.ShouldBindJSON(&newUser); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	user, err := c.userService.SignUp(newUser)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"accessToken": user.AccessToken, "user": user.User})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var loginUser domain.LoginDto
	if err := ctx.ShouldBindJSON(&loginUser); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	user, err := c.userService.Login(loginUser.Email, loginUser.Password)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"accessToken": user.AccessToken, "user": user.User})
}
