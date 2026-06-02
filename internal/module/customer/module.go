package customer

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/controllers"
	postgres "github.com/chand-magar/gamenglam-server/internal/module/customer/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/services"
)

type Module struct {
	CustomerHandler    *controllers.CustomerHandler
}

func NewModule(db *pgxpool.Pool) *Module {

	customerRepo := postgres.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepo)
	customerHandler := controllers.NewCustomerHandler(customerService)

	return &Module{
		CustomerHandler:    customerHandler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.CustomerRoutes(rg, m.CustomerHandler)
}