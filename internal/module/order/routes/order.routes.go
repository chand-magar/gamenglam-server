package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/order/controllers"
)

func OrderRoutes(rg *gin.RouterGroup, h *controllers.OrderHandler) {
	orders := rg.Group("/orders")
	{
		orders.GET("", h.FindAll)
		orders.POST("", h.PlaceOrder)
		orders.GET("/customer/:id", h.GetOrders)
		orders.GET("/:id", h.GetOrder)
		orders.PUT("/:id/status", h.UpdateStatus)
		orders.GET("/:id/history", h.GetStatusHistory)
	}
}
