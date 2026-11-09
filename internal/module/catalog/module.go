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
	VariantImageHandler          *controllers.VariantImageHandler
	ReviewHandler                *controllers.ReviewHandler
	productVariantService        services.ProductVariantService
}

func NewModule(db *pgxpool.Pool, baseURL string) *Module {
	categoryRepo := postgres.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := controllers.NewCategoryHandler(categoryService)

	productRepo := postgres.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := controllers.NewProductHandler(productService, baseURL)

	variantAttributeValueRepo := postgres.NewVariantAttributeValueRepository(db)
	variantAttributeValueService := services.NewVariantAttributeValueService(variantAttributeValueRepo)
	variantAttributeValueHandler := controllers.NewVariantAttributeValueHandler(variantAttributeValueService)

	productVariantRepo := postgres.NewProductVariantRepository(db)
	productVariantService := services.NewProductVariantService(productVariantRepo, variantAttributeValueRepo)
	productVariantHandler := controllers.NewProductVariantHandler(productVariantService, variantAttributeValueService)

	attributeRepo := postgres.NewAttributeRepository(db)
	attributeService := services.NewAttributeService(attributeRepo)
	attributeHandler := controllers.NewAttributeHandler(attributeService)

	attributeValueRepo := postgres.NewAttributeValueRepository(db)
	attributeValueService := services.NewAttributeValueService(attributeValueRepo)
	attributeValueHandler := controllers.NewAttributeValueHandler(attributeValueService)

	variantImageRepo := postgres.NewVariantImageRepository(db)
	variantImageHandler := controllers.NewVariantImageHandler(variantImageRepo, baseURL) // ← pass baseURL

	reviewRepo := postgres.NewReviewRepository(db)
	reviewService := services.NewReviewService(reviewRepo)
	reviewHandler := controllers.NewReviewHandler(reviewService)

	return &Module{
		CategoryHandler:              categoryHandler,
		ProductHandler:               productHandler,
		ProductVariantHandler:        productVariantHandler,
		AttributeHandler:             attributeHandler,
		AttributeValueHandler:        attributeValueHandler,
		VariantAttributeValueHandler: variantAttributeValueHandler,
		VariantImageHandler:          variantImageHandler,
		ReviewHandler:                reviewHandler,
		productVariantService:        productVariantService,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.CategoryRoutes(rg, m.CategoryHandler)
	routes.ProductRoutes(rg, m.ProductHandler, m.ProductVariantHandler)
	routes.ProductVariantRoutes(rg, m.ProductVariantHandler)
	routes.AttributeRoutes(rg, m.AttributeHandler, m.AttributeValueHandler)
	routes.AttributeValueRoutes(rg, m.AttributeValueHandler)
	routes.VariantAttributeValueRoutes(rg, m.VariantAttributeValueHandler)
	routes.VariantImageRoutes(rg, m.VariantImageHandler)
	routes.ReviewRoutes(rg, m.ReviewHandler)
}

func (m *Module) AsCatalogProvider() orderservices.CatalogProvider {
	return &catalogAdapter{svc: m.productVariantService}
}
