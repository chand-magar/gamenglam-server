package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/order/controllers"
)

func OrderRoutes(rg *gin.RouterGroup, h *controllers.OrderHandler) {
	orders := rg.Group("/orders")
	{
		orders.GET("",        h.FindAll)
		orders.GET("/:id",    h.FindOne)
		orders.POST("",       h.Create)
		orders.PUT("/:id",    h.Update)
		orders.DELETE("/:id", h.Delete)
	}
}