package catalog

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
	postgres "github.com/chand-magar/gamenglam-server/internal/module/catalog/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/services"
	orderservices "github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

type Module struct {
	CategoryHandler              *controllers.CategoryHandler
	ProductHandler               *controllers.ProductHandler
	ProductVariantHandler        *controllers.ProductVariantHandler
	AttributeHandler             *controllers.AttributeHandler
	AttributeValueHandler        *controllers.AttributeValueHandler
	VariantAttributeValueHandler *controllers.VariantAttributeValueHandler

	productVariantService services.ProductVariantService
}

func NewModule(db *pgxpool.Pool) *Module {
	categoryRepo := postgres.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := controllers.NewCategoryHandler(categoryService)

	productRepo := postgres.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := controllers.NewProductHandler(productService)

	productVariantRepo := postgres.NewProductVariantRepository(db)
	productVariantService := services.NewProductVariantService(productVariantRepo)
	productVariantHandler := controllers.NewProductVariantHandler(productVariantService)

	attributeRepo := postgres.NewAttributeRepository(db)
	attributeService := services.NewAttributeService(attributeRepo)
	attributeHandler := controllers.NewAttributeHandler(attributeService)

	attributeValueRepo := postgres.NewAttributeValueRepository(db)
	attributeValueService := services.NewAttributeValueService(attributeValueRepo)
	attributeValueHandler := controllers.NewAttributeValueHandler(attributeValueService)

	variantAttributeValueRepo := postgres.NewVariantAttributeValueRepository(db)
	variantAttributeValueService := services.NewVariantAttributeValueService(variantAttributeValueRepo)
	variantAttributeValueHandler := controllers.NewVariantAttributeValueHandler(variantAttributeValueService)

	return &Module{
		CategoryHandler:              categoryHandler,
		ProductHandler:               productHandler,
		ProductVariantHandler:        productVariantHandler,
		AttributeHandler:             attributeHandler,
		AttributeValueHandler:        attributeValueHandler,
		VariantAttributeValueHandler: variantAttributeValueHandler,
		productVariantService:        productVariantService,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.CategoryRoutes(rg, m.CategoryHandler)
	routes.ProductRoutes(rg, m.ProductHandler, m.ProductVariantHandler)
	routes.ProductVariantRoutes(rg, m.ProductVariantHandler, m.VariantAttributeValueHandler)
	routes.AttributeRoutes(rg, m.AttributeHandler, m.AttributeValueHandler)
	routes.AttributeValueRoutes(rg, m.AttributeValueHandler)
	routes.VariantAttributeValueRoutes(rg, m.VariantAttributeValueHandler)
}

func (m *Module) AsCatalogProvider() orderservices.CatalogProvider {
	return &catalogAdapter{svc: m.productVariantService}
}
