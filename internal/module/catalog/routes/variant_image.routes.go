package routes

import (
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/controllers"
	"github.com/gin-gonic/gin"
)

func VariantImageRoutes(rg *gin.RouterGroup, h *controllers.VariantImageHandler) {
	pv := rg.Group("/product-variants")
	{
		pv.GET("/:id/images", h.FindAll)
		pv.POST("/:id/images/upload", h.Upload)
		pv.POST("/:id/images/url", h.SaveURL)
	}

	vi := rg.Group("/variant-images")
	{
		vi.DELETE("/:id", h.Delete)
		vi.POST("/primary-batch", h.GetPrimaryBatch) // ← new
	}
}
