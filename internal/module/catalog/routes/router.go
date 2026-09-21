package routes

import (
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	CategoryHandler              *controllers.CategoryHandler
	ProductHandler               *controllers.ProductHandler
	ProductVariantHandler        *controllers.ProductVariantHandler
	AttributeHandler             *controllers.AttributeHandler
	AttributeValueHandler        *controllers.AttributeValueHandler
	VariantAttributeValueHandler *controllers.VariantAttributeValueHandler
}

func SetupRoutes(router *gin.Engine, deps Dependencies) {
	api := router.Group("/api/v1")
	CategoryRoutes(api, deps.CategoryHandler)
	ProductRoutes(api, deps.ProductHandler, deps.ProductVariantHandler)
	ProductVariantRoutes(api, deps.ProductVariantHandler) // ← removed extra arg
	AttributeRoutes(api, deps.AttributeHandler, deps.AttributeValueHandler)
	AttributeValueRoutes(api, deps.AttributeValueHandler)
	VariantAttributeValueRoutes(api, deps.VariantAttributeValueHandler)
}
