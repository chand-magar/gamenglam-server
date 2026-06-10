package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/order/controllers"
)

func OrderRoutes(rg *gin.RouterGroup, h *controllers.OrderHandler) {
	orders := rg.Group("/orders")
	{
		orders.POST("",                          h.PlaceOrder)
		orders.GET("/customer/:customer_id",     h.GetOrders)
		orders.GET("/:order_id",                 h.GetOrder)
		orders.PUT("/:order_id/status",          h.UpdateStatus)
		orders.GET("/:order_id/history",         h.GetStatusHistory)
	}
}