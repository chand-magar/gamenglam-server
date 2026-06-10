package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/controllers"
	postgres "github.com/chand-magar/gamenglam-server/internal/module/auth/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/services"
)

type Module struct {
	AuthHandler *controllers.AuthHandler
}

func NewModule(db *pgxpool.Pool) *Module {
	// User
	userRepo := postgres.NewUserRepository(db)

	// Auth
	credRepo := postgres.NewCredentialRepository(db)
	authService := services.NewAuthService(credRepo, userRepo)
	authHandler := controllers.NewAuthHandler(authService)

	return &Module{
		AuthHandler: authHandler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {

	routes.AuthRoutes(rg, m.AuthHandler)
}
