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
	cartHandler   *controllers.CartHandler
	orderHandler  *controllers.OrderHandler
	couponHandler *controllers.CouponHandler
}

func NewModule(db *pgxpool.Pool, catalogProvider services.CatalogProvider, customerProvider services.CustomerProvider) *Module {
	// Repositories
	cartRepo   := postgres.NewCartRepository(db)
	orderRepo  := postgres.NewOrderRepository(db)
	couponRepo := postgres.NewCouponRepository(db)

	// Services
	cartService   := services.NewCartService(cartRepo, catalogProvider)
	orderService  := services.NewOrderService(orderRepo, catalogProvider, customerProvider)
	couponService := services.NewCouponService(couponRepo)

	// Controllers
	cartHandler   := controllers.NewCartHandler(cartService)
	orderHandler  := controllers.NewOrderHandler(orderService)
	couponHandler := controllers.NewCouponHandler(couponService)

	return &Module{
		cartHandler:   cartHandler,
		orderHandler:  orderHandler,
		couponHandler: couponHandler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.CartRoutes(rg, m.cartHandler)
	routes.OrderRoutes(rg, m.orderHandler)
	routes.CouponRoutes(rg, m.couponHandler)
}