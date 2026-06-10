package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/order/controllers"
)

func CartRoutes(rg *gin.RouterGroup, h *controllers.CartHandler) {
	cart := rg.Group("/cart")
	{
		cart.GET("/:customer_id",            h.GetCart)
		cart.POST("/:customer_id/items",     h.AddItem)
		cart.PUT("/items/:cart_item_id",     h.UpdateItem)
		cart.DELETE("/items/:cart_item_id",  h.RemoveItem)
		cart.DELETE("/:customer_id/clear",   h.ClearCart)
	}
}