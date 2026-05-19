package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/user"
)

func SetupRouter(router *gin.Engine, db *pgxpool.Pool) {
	api := router.Group("/api/v1")

	userModule := user.NewModule(db)
	userModule.RegisterRoutes(api)
}