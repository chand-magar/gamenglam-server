package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
)

func VariantAttributeValueRoutes(rg *gin.RouterGroup, h *controllers.VariantAttributeValueHandler) {
	variantAttributeValues := rg.Group("/variant-attribute-values")
	{
		variantAttributeValues.GET("",        h.FindAll)
		variantAttributeValues.GET("/:id",    h.FindOne)
		variantAttributeValues.POST("",       h.Create)
		// variantAttributeValues.PUT("/:id",    h.Update)
		variantAttributeValues.DELETE("/:id", h.Delete)
	}
}