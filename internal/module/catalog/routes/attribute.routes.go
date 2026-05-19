package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
)

func AttributeRoutes(rg *gin.RouterGroup, h *controllers.AttributeHandler) {
	attributes := rg.Group("/attributes")
	{
		attributes.GET("",        h.FindAll)
		attributes.GET("/:id",    h.FindOne)
		attributes.POST("",       h.Create)
		attributes.PUT("/:id",    h.Update)
		attributes.DELETE("/:id", h.Delete)
	}
}