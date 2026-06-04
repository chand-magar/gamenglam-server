package order

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/order/controllers"
	"github.com/chand-magar/gamenglam-server/internal/module/order/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/order/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

type Module struct {
	cartHandler  *controllers.CartHandler
	orderHandler *controllers.OrderHandler
}

func NewModule(db *pgxpool.Pool, catalogProvider services.CatalogProvider, customerProvider services.CustomerProvider) *Module {
	// Repositories
	cartRepo  := postgres.NewCartRepository(db)
	orderRepo := postgres.NewOrderRepository(db)

	// Services
	cartService  := services.NewCartService(cartRepo, catalogProvider)
	orderService := services.NewOrderService(orderRepo, catalogProvider, customerProvider)

	// Controllers
	cartHandler  := controllers.NewCartHandler(cartService)
	orderHandler := controllers.NewOrderHandler(orderService)

	return &Module{
		cartHandler:  cartHandler,
		orderHandler: orderHandler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.RegisterRoutes(rg, m.cartHandler, m.orderHandler)
}