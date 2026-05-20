package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool" // pgxpool, not gorm

	"github.com/chand-magar/gamenglam-server/internal/module/auth/controllers"
	postgres "github.com/chand-magar/gamenglam-server/internal/module/auth/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/auth/services"
)

type Module struct {
	UserHandler    *controllers.UserHandler
	RoleHandler    *controllers.RoleHandler
	SectionHandler *controllers.SectionHandler
	PageHandler    *controllers.PageHandler
}

func NewModule(db *pgxpool.Pool) *Module {

	userRepo := postgres.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := controllers.NewUserHandler(userService)

	roleRepo := postgres.NewRoleRepository(db)
	roleService := services.NewRoleService(roleRepo)
	roleHandler := controllers.NewRoleHandler(roleService)

	sectionRepo := postgres.NewSectionRepository(db)
	sectionService := services.NewSectionService(sectionRepo)
	sectionHandler := controllers.NewSectionHandler(sectionService)

	pageRepo := postgres.NewPageRepository(db)
	pageService := services.NewPageService(pageRepo)
	pageHandler := controllers.NewPageHandler(pageService)

	return &Module{
		UserHandler:    userHandler,
		RoleHandler:    roleHandler,
		SectionHandler: sectionHandler,
		PageHandler:    pageHandler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.UserRoutes(rg, m.UserHandler)
	routes.RoleRoutes(rg, m.RoleHandler)
	routes.SectionRoutes(rg, m.SectionHandler)
	routes.PageRoutes(rg, m.PageHandler)
}