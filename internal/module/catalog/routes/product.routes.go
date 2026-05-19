package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
)

func ProductRoutes(rg *gin.RouterGroup, h *controllers.ProductHandler) {
	products := rg.Group("/products")
	{
		products.GET("",        h.FindAll)
		products.GET("/:id",    h.FindOne)
		products.POST("",       h.Create)
		products.PUT("/:id",    h.Update)
		products.DELETE("/:id", h.Delete)
	}
}