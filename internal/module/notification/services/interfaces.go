package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/models"
)

type SubscriberService interface {
	GetAll(ctx context.Context, req *models.SubscriberListRequest) (*models.SubscriberListResponse, error)
	Subscribe(ctx context.Context, req *models.CreateSubscriberRequest) (*models.Subscriber, error)
	Unsubscribe(ctx context.Context, id uuid.UUID) error
}

type CampaignService interface {
	GetAll(ctx context.Context, page, limit int) ([]models.Campaign, int, int, error)
	Send(ctx context.Context, req *models.SendCampaignRequest) (*models.Campaign, error)
}
