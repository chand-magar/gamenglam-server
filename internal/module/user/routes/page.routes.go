package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/user/controllers"
)

func PageRoutes(rg *gin.RouterGroup, h *controllers.PageHandler) {
	pages := rg.Group("/pages")
	{
		pages.GET("", h.FindAll)
		pages.GET("/:id", h.FindOne)
		pages.POST("", h.Create)
		pages.PUT("/:id", h.Update)
		pages.DELETE("/:id", h.Delete)
	}
}