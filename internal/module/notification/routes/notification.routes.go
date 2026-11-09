package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/controllers"
)

func NewsletterRoutes(rg *gin.RouterGroup, sub *controllers.SubscriberHandler, camp *controllers.CampaignHandler) {
	nl := rg.Group("/newsletter")
	{
		// Public — storefront subscribe form
		nl.POST("/subscribe", sub.Subscribe)

		// Admin — subscriber management
		subscribers := nl.Group("/subscribers")
		{
			subscribers.GET("",       sub.FindAll)
			subscribers.POST("",      sub.AddSubscriber)
			subscribers.DELETE("/:id", sub.Unsubscribe)
		}

		// Admin — campaign management
		campaigns := nl.Group("/campaigns")
		{
			campaigns.GET("",     camp.FindAll)
			campaigns.POST("/send", camp.Send)
		}
	}
}
