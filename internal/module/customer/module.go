package customer

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/controllers"
	postgres "github.com/chand-magar/gamenglam-server/internal/module/customer/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/services"
	orderservices "github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

type Module struct {
	CustomerHandler *controllers.CustomerHandler

	// exposed for cross-module adapter
	addressService services.AddressService
}

func NewModule(db *pgxpool.Pool) *Module {
	customerRepo    := postgres.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepo)
	customerHandler := controllers.NewCustomerHandler(customerService)

	addressRepo    := postgres.NewAddressRepository(db)
	addressService := services.NewAddressService(addressRepo)

	return &Module{
		CustomerHandler: customerHandler,
		addressService:  addressService,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.CustomerRoutes(rg, m.CustomerHandler)
}

func (m *Module) AsCustomerProvider() orderservices.CustomerProvider {
	return &customerAdapter{svc: m.addressService}
}