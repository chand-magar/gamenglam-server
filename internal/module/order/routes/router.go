package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/order/controllers"
)

func RegisterRoutes(rg *gin.RouterGroup, cartHandler *controllers.CartHandler, orderHandler *controllers.OrderHandler) {
	// Cart routes
	cart := rg.Group("/cart")
	{
		cart.GET("/:customer_no",                    cartHandler.GetCart)
		cart.POST("/:customer_no/items",             cartHandler.AddItem)
		cart.PUT("/items/:cart_item_id",             cartHandler.UpdateItem)
		cart.DELETE("/items/:cart_item_id",          cartHandler.RemoveItem)
		cart.DELETE("/:customer_no/clear",           cartHandler.ClearCart)
	}

	// Order routes
	orders := rg.Group("/orders")
	{
		orders.POST("",                              orderHandler.PlaceOrder)
		orders.GET("/customer/:customer_no",         orderHandler.GetOrders)
		orders.GET("/:order_id",                     orderHandler.GetOrder)
		orders.PUT("/:order_id/status",              orderHandler.UpdateStatus)
		orders.GET("/:order_no/history",             orderHandler.GetStatusHistory)
	}
}