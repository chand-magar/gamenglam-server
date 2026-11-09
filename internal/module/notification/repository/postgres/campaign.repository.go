package postgres

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/models"
	repository "github.com/chand-magar/gamenglam-server/internal/module/notification/repository"
)

type campaignRepository struct {
	db *pgxpool.Pool
}

func NewCampaignRepository(db *pgxpool.Pool) repository.CampaignRepository {
	return &campaignRepository{db: db}
}

func (r *campaignRepository) GetAll(ctx context.Context, page, limit int) (*models.SubscriberListResponse, error) {
	return nil, nil // unused — GetAllCampaigns is used directly
}

func (r *campaignRepository) GetAllCampaigns(ctx context.Context, page, limit int) ([]models.Campaign, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notification.newsletter_campaigns`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("campaignRepository.GetAllCampaigns count: %w", err)
	}

	_ = int(math.Ceil(float64(total) / float64(limit))) // totalPages unused here, returned by service

	rows, err := r.db.Query(ctx, `
		SELECT campaign_no, campaign_id, subject, body, sent_status, recipient_count, sent_at,
		       COALESCE(error_message,''), created_at
		FROM notification.newsletter_campaigns
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("campaignRepository.GetAllCampaigns query: %w", err)
	}
	defer rows.Close()

	var items []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(
			&c.CampaignNo, &c.CampaignId, &c.Subject, &c.Body, &c.SentStatus,
			&c.RecipientCount, &c.SentAt, &c.ErrorMessage, &c.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("campaignRepository.GetAllCampaigns scan: %w", err)
		}
		items = append(items, c)
	}
	if items == nil {
		items = []models.Campaign{}
	}
	return items, total, nil
}

func (r *campaignRepository) Save(ctx context.Context, campaign *models.Campaign) (*models.Campaign, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO notification.newsletter_campaigns (subject, body, sent_status)
		VALUES ($1, $2, 'draft')
		RETURNING campaign_no, campaign_id, subject, body, sent_status, recipient_count, sent_at,
		          COALESCE(error_message,''), created_at
	`, campaign.Subject, campaign.Body).Scan(
		&campaign.CampaignNo, &campaign.CampaignId, &campaign.Subject, &campaign.Body,
		&campaign.SentStatus, &campaign.RecipientCount, &campaign.SentAt,
		&campaign.ErrorMessage, &campaign.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("campaignRepository.Save: %w", err)
	}
	return campaign, nil
}

func (r *campaignRepository) MarkSent(ctx context.Context, id uuid.UUID, count int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE notification.newsletter_campaigns
		SET sent_status = 'sent', recipient_count = $2, sent_at = NOW(), updated_at = NOW()
		WHERE campaign_id = $1
	`, id, count)
	if err != nil {
		return fmt.Errorf("campaignRepository.MarkSent: %w", err)
	}
	return nil
}

func (r *campaignRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE notification.newsletter_campaigns
		SET sent_status = 'failed', error_message = $2, updated_at = NOW()
		WHERE campaign_id = $1
	`, id, errMsg)
	if err != nil {
		return fmt.Errorf("campaignRepository.MarkFailed: %w", err)
	}
	return nil
}
