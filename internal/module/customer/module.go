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
	CustomerHandler     *controllers.CustomerHandler
	CustomerAuthHandler *controllers.CustomerAuthHandler
	AddressHandler      *controllers.AddressHandler

	addressService services.AddressService // ← store for reuse
}

func NewModule(db *pgxpool.Pool) *Module {
	customerRepo := postgres.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepo)
	customerHandler := controllers.NewCustomerHandler(customerService)

	addressRepo := postgres.NewAddressRepository(db)
	addressService := services.NewAddressService(addressRepo)
	addressHandler := controllers.NewAddressHandler(addressService, customerService)

	credRepo := postgres.NewCustomerCredentialRepository(db)
	customerAuthSvc := services.NewCustomerAuthService(customerRepo, credRepo)
	customerAuthHandler := controllers.NewCustomerAuthHandler(customerAuthSvc)

	return &Module{
		CustomerHandler:     customerHandler,
		CustomerAuthHandler: customerAuthHandler,
		AddressHandler:      addressHandler,
		addressService:      addressService, // ← store real service
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.CustomerRoutes(rg, m.CustomerHandler)
	routes.CustomerAuthRoutes(rg, m.CustomerAuthHandler)
	routes.AddressRoutes(rg, m.AddressHandler)
}

func (m *Module) AsCustomerProvider() orderservices.CustomerProvider {
	return &customerAdapter{svc: m.addressService} // ← use real service, not nil
}
