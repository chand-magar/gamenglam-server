package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
)

func CategoryRoutes(rg *gin.RouterGroup, h *controllers.CategoryHandler) {
	categories := rg.Group("/categories")
	{
		categories.GET("",        h.FindAll)
		categories.GET("/:id",    h.FindOne)
		categories.POST("",       h.Create)
		categories.PUT("/:id",    h.Update)
		categories.DELETE("/:id", h.Delete)
	}
}