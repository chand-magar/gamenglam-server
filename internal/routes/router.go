package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/config"
	"github.com/chand-magar/gamenglam-server/internal/module/auth"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog"
	"github.com/chand-magar/gamenglam-server/internal/module/customer"
	"github.com/chand-magar/gamenglam-server/internal/module/employee"
	"github.com/chand-magar/gamenglam-server/internal/module/notification"
	"github.com/chand-magar/gamenglam-server/internal/module/order"
	"github.com/chand-magar/gamenglam-server/internal/module/stats"
)

func SetupRouter(router *gin.Engine, db *pgxpool.Pool, baseURL string, cfg *config.Config) {
	api := router.Group("/api/v1")

	authModule := auth.NewModule(db)
	authModule.RegisterRoutes(api)

	employeeModule := employee.NewModule(db)
	employeeModule.RegisterRoutes(api)

	catalogModule := catalog.NewModule(db, baseURL)
	catalogModule.RegisterRoutes(api)

	customerModule := customer.NewModule(db)
	customerModule.RegisterRoutes(api)

	orderModule := order.NewModule(
		db,
		catalogModule.AsCatalogProvider(),
		customerModule.AsCustomerProvider(),
	)
	orderModule.RegisterRoutes(api)

	notificationModule := notification.NewModule(db, cfg.SMTP)
	notificationModule.RegisterRoutes(api)

	statsHandler := stats.NewHandler(db)
	api.GET("/stats", statsHandler.GetStats)
}
