package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/models"
)

type SubscriberRepository interface {
	GetAll(ctx context.Context, req *models.SubscriberListRequest) (*models.SubscriberListResponse, error)
	GetActiveEmails(ctx context.Context) ([]string, error)
	Save(ctx context.Context, req *models.CreateSubscriberRequest) (*models.Subscriber, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type CampaignRepository interface {
	GetAll(ctx context.Context, page, limit int) (*models.SubscriberListResponse, error)
	GetAllCampaigns(ctx context.Context, page, limit int) ([]models.Campaign, int, error)
	Save(ctx context.Context, campaign *models.Campaign) (*models.Campaign, error)
	MarkSent(ctx context.Context, id uuid.UUID, count int) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}
