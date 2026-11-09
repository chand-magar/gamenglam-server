package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/order/controllers"
)

func CouponRoutes(rg *gin.RouterGroup, h *controllers.CouponHandler) {
	coupons := rg.Group("/coupons")
	{
		coupons.GET("",     h.FindAll)
		coupons.GET("/:id", h.FindOne)
		coupons.POST("",    h.Create)
		coupons.PUT("/:id", h.Update)
		coupons.DELETE("/:id", h.Delete)
	}
}
