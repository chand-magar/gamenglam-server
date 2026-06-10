package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/employee/controllers"
)

func RoleRoutes(rg *gin.RouterGroup, h *controllers.RoleHandler) {
	roles := rg.Group("/roles")
	{
		roles.GET("", h.FindAll)
		roles.GET("/:id", h.FindOne)
		roles.POST("", h.Create)
		roles.PUT("/:id", h.Update)
		roles.DELETE("/:id", h.Delete)
	}
}
