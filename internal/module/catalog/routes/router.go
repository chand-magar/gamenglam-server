package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
)

type Dependencies struct {
	CategoryHandler *controllers.CategoryHandler
	ProductHandler *controllers.ProductHandler
	ProductVariantHandler *controllers.ProductVariantHandler
	AttributeHandler *controllers.AttributeHandler
	AttributeValueHandler *controllers.AttributeValueHandler
	VariantAttributeValueHandler *controllers.VariantAttributeValueHandler
}

func SetupRoutes(router *gin.Engine, deps Dependencies) {
	api := router.Group("/api/v1")
	CategoryRoutes(api, deps.CategoryHandler)
	ProductRoutes(api, deps.ProductHandler)
	ProductVariantRoutes(api, deps.ProductVariantHandler)
	AttributeRoutes(api, deps.AttributeHandler)
	AttributeValueRoutes(api, deps.AttributeValueHandler)
	VariantAttributeValueRoutes(api, deps.VariantAttributeValueHandler)
}