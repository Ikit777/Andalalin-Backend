package routes

import (
	"github.com/Ikit777/Andalalin-Backend/controllers"
	"github.com/Ikit777/Andalalin-Backend/middleware"
	"github.com/gin-gonic/gin"
)

type AuthRouteController struct {
	authController controllers.AuthController
}

func NewAuthRouteController(authController controllers.AuthController) AuthRouteController {
	return AuthRouteController{authController}
}

func (rc *AuthRouteController) AuthRoute(rg *gin.RouterGroup) {
	router := rg.Group("/auth")

	router.POST("/register", rc.authController.SignUp)

	router.POST("/login", rc.authController.SignIn)

	router.GET("/refresh", rc.authController.RefreshAccessToken)

	router.GET("/verification/:verificationCode", rc.authController.VerifyEmail)
	router.POST("/resend", rc.authController.ResendVerification)

	router.GET("/logout", middleware.DeserializeUser(), rc.authController.LogoutUser)
}
