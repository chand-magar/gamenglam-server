package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
)

func ProductVariantRoutes(
	rg *gin.RouterGroup,
	h *controllers.ProductVariantHandler,
	vh *controllers.VariantAttributeValueHandler,
) {
	productVariant := rg.Group("/product-variants")
	{
		productVariant.GET("", h.FindAll)
		productVariant.GET("/:id", h.FindOne)
		productVariant.POST("", h.Create)
		productVariant.PUT("/:id", h.Update)
		productVariant.DELETE("/:id", h.Delete)
		productVariant.GET("/:id/attribute-values", vh.FindByVariantId)
	}
}
