package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
)

func ReviewRoutes(rg *gin.RouterGroup, h *controllers.ReviewHandler) {
	products := rg.Group("/products")
	{
		products.GET("/:id/reviews", h.FindByProduct)
		products.POST("/:id/reviews", h.Create)
	}

	// Admin routes
	reviews := rg.Group("/reviews")
	{
		reviews.GET("",      h.FindAll)
		reviews.DELETE("/:id", h.DeleteOne)
	}
}
