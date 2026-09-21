package routes

import (
	"github.com/chand-magar/gamenglam-server/internal/module/order/controllers"
	"github.com/gin-gonic/gin"
)

func CartRoutes(rg *gin.RouterGroup, h *controllers.CartHandler) {
	cart := rg.Group("/cart")
	{
		cart.GET("/:id", h.GetCart)
		cart.POST("/:id/items", h.AddItem)
		cart.PUT("/items/:id", h.UpdateItem)
		cart.DELETE("/items/:id", h.RemoveItem)
		cart.DELETE("/:id/clear", h.ClearCart)
	}
}
