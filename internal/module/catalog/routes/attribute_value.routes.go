package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
)

func AttributeValueRoutes(rg *gin.RouterGroup, h *controllers.AttributeValueHandler) {
	attributeValues := rg.Group("/attribute-values")
	{
		attributeValues.GET("", h.FindAll)
		attributeValues.GET("/:id", h.FindOne)
		attributeValues.POST("", h.Create)
		attributeValues.PUT("/:id", h.Update)
		attributeValues.DELETE("/:id", h.Delete)
	}
}
