package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/auth/controllers"
)

func UserRoutes(rg *gin.RouterGroup, h *controllers.UserHandler) {
	users := rg.Group("/users")
	{
		users.GET("",        h.FindAll)
		users.GET("/:id",    h.FindOne)
		users.POST("",       h.Create)
		users.PUT("/:id",    h.Update)
		users.DELETE("/:id", h.Delete)
	}
}