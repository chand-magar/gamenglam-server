package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/controllers"
)

func CustomerRoutes(rg *gin.RouterGroup, h *controllers.CustomerHandler) {
	customers := rg.Group("/customers")
	{
		// Admin CRUD
		customers.GET("", h.FindAll)
		customers.GET("/:id", h.FindOne)
		customers.POST("", h.Create)
		customers.PUT("/:id", h.Update)
		customers.DELETE("/:id", h.Delete)
	}
}

func CustomerAuthRoutes(rg *gin.RouterGroup, h *controllers.CustomerAuthHandler) {
	customers := rg.Group("/customers")
	{
		// Public — no auth required
		customers.POST("/register", h.Register)
		customers.POST("/login", h.Login)
		customers.POST("/refresh", h.RefreshToken)
		customers.POST("/logout", h.Logout)

		// Protected — requires customer JWT
		customers.GET("/profile", h.GetProfile)
	}
}
