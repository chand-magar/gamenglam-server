package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/auth"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog"
	"github.com/chand-magar/gamenglam-server/internal/module/customer"
	"github.com/chand-magar/gamenglam-server/internal/module/order"
	orderservices "github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

func SetupRouter(router *gin.Engine, db *pgxpool.Pool) {
	api := router.Group("/api/v1")

	authModule := auth.NewModule(db)
	authModule.RegisterRoutes(api)

	catalogModule := catalog.NewModule(db)
	catalogModule.RegisterRoutes(api)

	customerModule := customer.NewModule(db)
	customerModule.RegisterRoutes(api)

	// Adapters — wire catalog and customer into order module
	orderModule := order.NewModule(
		db,
		catalogModule.AsCatalogProvider(),
		customerModule.AsCustomerProvider(),
	)
	orderModule.RegisterRoutes(api)
}