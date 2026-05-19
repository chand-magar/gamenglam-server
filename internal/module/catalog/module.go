package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
	postgres "github.com/chand-magar/gamenglam-server/internal/module/catalog/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/services"
)

type Module struct {
	UserHandler    *controllers.UserHandler
	RoleHandler    *controllers.RoleHandler
	SectionHandler *controllers.SectionHandler
	PageHandler    *controllers.PageHandler
}

func NewModule(db *pgxpool.Pool) *Module {

	categoryRepo    := postgresrepo.NewCategoryRepository(db)
	categryService 	:= services.NewCategoryService(categoryRepo)
	categoryHandler := controllers.NewCategoryHandler(categryService)

	productRepo    := postgresrepo.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := controllers.NewProductHandler(productService)

	productVariantRepo    := postgresrepo.NewProductVariantRepository(db)
	productVariantService := services.NewProductVariantService(productVariantRepo)
	productVariantHandler := controllers.NewProductVariantHandler(productVariantService)

	attributeRepo    := postgresrepo.NewAttributeRepository(db)
	attributervice := services.NewAttributeService(attributeRepo)
	attributeHandler := controllers.NewAttributeHandler(attributervice)

	attributeValueRepo    := postgresrepo.NewAttributeValueRepository(db)
	attributeValueService := services.NewAttributeValueService(attributeValueRepo)
	attributeValueHandler := controllers.NewAttributeValueHandler(attributeValueService)

	variantAttributeValueRepo    := postgresrepo.NewVariantAttributeValueRepository(db)
	variantAttributeValueService := services.NewVariantAttributeValueService(variantAttributeValueRepo)
	variantAttributeValueHandler := controllers.NewVariantAttributeValueHandler(variantAttributeValueService)

	return &Module{
		CategoryHandler:    			categoryHandler,
		ProductHandler:    				productHandler,
		ProductVariantHandler: 			productVariantHandler,
		AttributeHandler:    			attributeHandler,
		AttributeValueHandler:			attributeValueHandler,
		VariantAttributeValueHandler: 	variantAttributeValueHandler
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