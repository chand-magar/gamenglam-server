package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/controllers"
)

func AddressRoutes(rg *gin.RouterGroup, h *controllers.AddressHandler) {
	addresses := rg.Group("/addresses")
	{
		addresses.GET("", h.FindAll)       // GET    /api/v1/addresses
		addresses.POST("", h.Create)       // POST   /api/v1/addresses
		addresses.PUT("/:id", h.Update)    // PUT    /api/v1/addresses/:id
		addresses.DELETE("/:id", h.Delete) // DELETE /api/v1/addresses/:id
	}
}
