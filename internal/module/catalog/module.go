package catalog

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
	postgres "github.com/chand-magar/gamenglam-server/internal/module/catalog/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/services"
)

type Module struct {
	CategoryHandler    *controllers.CategoryHandler
	ProductHandler    *controllers.ProductHandler
	ProductVariantHandler *controllers.ProductVariantHandler
	AttributeHandler    *controllers.AttributeHandler
	AttributeValueHandler *controllers.AttributeValueHandler
	VariantAttributeValueHandler *controllers.VariantAttributeValueHandler
}

func NewModule(db *pgxpool.Pool) *Module {

	categoryRepo    := postgres.NewCategoryRepository(db)
	categryService 	:= services.NewCategoryService(categoryRepo)
	categoryHandler := controllers.NewCategoryHandler(categryService)

	productRepo    := postgres.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := controllers.NewProductHandler(productService)

	productVariantRepo    := postgres.NewProductVariantRepository(db)
	productVariantService := services.NewProductVariantService(productVariantRepo)
	productVariantHandler := controllers.NewProductVariantHandler(productVariantService)

	attributeRepo    := postgres.NewAttributeRepository(db)
	attributervice := services.NewAttributeService(attributeRepo)
	attributeHandler := controllers.NewAttributeHandler(attributervice)

	attributeValueRepo    := postgres.NewAttributeValueRepository(db)
	attributeValueService := services.NewAttributeValueService(attributeValueRepo)
	attributeValueHandler := controllers.NewAttributeValueHandler(attributeValueService)

	variantAttributeValueRepo    := postgres.NewVariantAttributeValueRepository(db)
	variantAttributeValueService := services.NewVariantAttributeValueService(variantAttributeValueRepo)
	variantAttributeValueHandler := controllers.NewVariantAttributeValueHandler(variantAttributeValueService)

	return &Module{
		CategoryHandler:    			categoryHandler,
		ProductHandler:    				productHandler,
		ProductVariantHandler: 			productVariantHandler,
		AttributeHandler:    			attributeHandler,
		AttributeValueHandler:			attributeValueHandler,
		VariantAttributeValueHandler: 	variantAttributeValueHandler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.CategoryRoutes(rg, m.CategoryHandler)
	routes.ProductRoutes(rg, m.ProductHandler)
	routes.ProductVariantRoutes(rg, m.ProductVariantHandler)
	routes.AttributeRoutes(rg, m.AttributeHandler)
	routes.AttributeValueRoutes(rg, m.AttributeValueHandler)
	routes.VariantAttributeValueRoutes(rg, m.VariantAttributeValueHandler)
}