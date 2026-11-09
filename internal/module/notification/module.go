package notification

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/config"
	"github.com/chand-magar/gamenglam-server/internal/module/notification/controllers"
	postgres "github.com/chand-magar/gamenglam-server/internal/module/notification/repository/postgres"
	"github.com/chand-magar/gamenglam-server/internal/module/notification/routes"
	"github.com/chand-magar/gamenglam-server/internal/module/notification/services"
)

type Module struct {
	subscriberHandler *controllers.SubscriberHandler
	campaignHandler   *controllers.CampaignHandler
}

func NewModule(db *pgxpool.Pool, smtpCfg config.SMTP) *Module {
	subscriberRepo := postgres.NewSubscriberRepository(db)
	campaignRepo   := postgres.NewCampaignRepository(db)

	mailer := services.NewMailer(services.MailerConfig{
		Host:      smtpCfg.Host,
		Port:      smtpCfg.Port,
		Username:  smtpCfg.Username,
		Password:  smtpCfg.Password,
		FromName:  smtpCfg.FromName,
		FromEmail: smtpCfg.FromEmail,
	})

	subscriberService := services.NewSubscriberService(subscriberRepo)
	campaignService   := services.NewCampaignService(campaignRepo, subscriberRepo, mailer)

	return &Module{
		subscriberHandler: controllers.NewSubscriberHandler(subscriberService),
		campaignHandler:   controllers.NewCampaignHandler(campaignService),
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	routes.NewsletterRoutes(rg, m.subscriberHandler, m.campaignHandler)
}
