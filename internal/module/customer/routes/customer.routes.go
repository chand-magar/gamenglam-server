package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/controllers"
)

func CustomerRoutes(rg *gin.RouterGroup, h *controllers.CustomerHandler) {
	customers := rg.Group("/customers")
	{
		customers.GET("",        h.FindAll)
		customers.GET("/:id",    h.FindOne)
		customers.POST("",       h.Create)
		customers.PUT("/:id",    h.Update)
		customers.DELETE("/:id", h.Delete)
	}
}