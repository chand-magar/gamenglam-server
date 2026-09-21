package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/auth"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog"
	"github.com/chand-magar/gamenglam-server/internal/module/customer"
	"github.com/chand-magar/gamenglam-server/internal/module/employee"
	"github.com/chand-magar/gamenglam-server/internal/module/order"
)

func SetupRouter(router *gin.Engine, db *pgxpool.Pool, baseURL string) {
	api := router.Group("/api/v1")

	authModule := auth.NewModule(db)
	authModule.RegisterRoutes(api)

	employeeModule := employee.NewModule(db)
	employeeModule.RegisterRoutes(api)

	catalogModule := catalog.NewModule(db, baseURL) // ← pass baseURL
	catalogModule.RegisterRoutes(api)

	customerModule := customer.NewModule(db)
	customerModule.RegisterRoutes(api)

	orderModule := order.NewModule(
		db,
		catalogModule.AsCatalogProvider(),
		customerModule.AsCustomerProvider(),
	)
	orderModule.RegisterRoutes(api)
}
