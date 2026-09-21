package routes

import (
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
	"github.com/gin-gonic/gin"
)

func ProductVariantRoutes(rg *gin.RouterGroup, h *controllers.ProductVariantHandler) {
	productVariant := rg.Group("/product-variants")
	{
		productVariant.GET("", h.FindAll)
		productVariant.GET("/listing", h.GetListing) // ← add before /:id
		productVariant.GET("/:id", h.FindOne)
		productVariant.POST("", h.Create)
		productVariant.POST("/bulk", h.BulkCreate)
		productVariant.PUT("/:id", h.Update)
		productVariant.DELETE("/:id", h.Delete)
		productVariant.GET("/:id/attribute-values", h.GetAttributeValues) // ← add
	}
}
