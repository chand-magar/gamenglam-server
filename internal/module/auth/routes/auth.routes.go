package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/controllers"
)

func AuthRoutes(rg *gin.RouterGroup, h *controllers.AuthHandler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh-token", h.RefreshToken)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/verify-otp", h.VerifyOTP)
		auth.POST("/reset-password", h.ResetPassword)
		auth.POST("/change-password", h.ChangePassword)
		auth.GET("/me", h.Me)
	}
}
