package services

import (
	"context"
	"fmt"
	"math"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/models"
	repository "github.com/chand-magar/gamenglam-server/internal/module/notification/repository"
)

type campaignService struct {
	campaignRepo   repository.CampaignRepository
	subscriberRepo repository.SubscriberRepository
	mailer         Mailer
}

func NewCampaignService(
	campaignRepo repository.CampaignRepository,
	subscriberRepo repository.SubscriberRepository,
	mailer Mailer,
) CampaignService {
	return &campaignService{
		campaignRepo:   campaignRepo,
		subscriberRepo: subscriberRepo,
		mailer:         mailer,
	}
}

func (s *campaignService) GetAll(ctx context.Context, page, limit int) ([]models.Campaign, int, int, error) {
	items, total, err := s.campaignRepo.GetAllCampaigns(ctx, page, limit)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("campaignService.GetAll: %w", err)
	}
	if limit < 1 {
		limit = 10
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return items, total, totalPages, nil
}

func (s *campaignService) Send(ctx context.Context, req *models.SendCampaignRequest) (*models.Campaign, error) {
	// Persist campaign record first
	campaign, err := s.campaignRepo.Save(ctx, &models.Campaign{
		Subject: req.Subject,
		Body:    req.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("campaignService.Send: persist: %w", err)
	}

	if !s.mailer.Configured() {
		_ = s.campaignRepo.MarkFailed(ctx, campaign.CampaignId, "SMTP not configured")
		campaign.SentStatus = "failed"
		campaign.ErrorMessage = "SMTP not configured — set SMTP credentials in server config"
		return campaign, nil
	}

	emails, err := s.subscriberRepo.GetActiveEmails(ctx)
	if err != nil {
		_ = s.campaignRepo.MarkFailed(ctx, campaign.CampaignId, err.Error())
		return nil, fmt.Errorf("campaignService.Send: get emails: %w", err)
	}

	if len(emails) == 0 {
		_ = s.campaignRepo.MarkFailed(ctx, campaign.CampaignId, "no active subscribers")
		campaign.SentStatus = "failed"
		campaign.ErrorMessage = "No active subscribers"
		return campaign, nil
	}

	if err := s.mailer.Send(emails, req.Subject, req.Body); err != nil {
		_ = s.campaignRepo.MarkFailed(ctx, campaign.CampaignId, err.Error())
		campaign.SentStatus = "failed"
		campaign.ErrorMessage = err.Error()
		return campaign, nil
	}

	if err := s.campaignRepo.MarkSent(ctx, campaign.CampaignId, len(emails)); err != nil {
		return nil, fmt.Errorf("campaignService.Send: mark sent: %w", err)
	}
	campaign.SentStatus = "sent"
	campaign.RecipientCount = len(emails)
	return campaign, nil
}
