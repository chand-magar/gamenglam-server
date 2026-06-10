package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/employee/controllers"
)

func SectionRoutes(rg *gin.RouterGroup, h *controllers.SectionHandler) {
	sections := rg.Group("/sections")
	{
		sections.GET("", h.FindAll)
		sections.GET("/:id", h.FindOne)
		sections.POST("", h.Create)
		sections.PUT("/:id", h.Update)
		sections.DELETE("/:id", h.Delete)
	}
}
