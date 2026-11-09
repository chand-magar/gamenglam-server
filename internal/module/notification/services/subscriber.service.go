package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/notification/models"
	repository "github.com/chand-magar/gamenglam-server/internal/module/notification/repository"
)

type subscriberService struct {
	repo repository.SubscriberRepository
}

func NewSubscriberService(repo repository.SubscriberRepository) SubscriberService {
	return &subscriberService{repo: repo}
}

func (s *subscriberService) GetAll(ctx context.Context, req *models.SubscriberListRequest) (*models.SubscriberListResponse, error) {
	result, err := s.repo.GetAll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("subscriberService.GetAll: %w", err)
	}
	return result, nil
}

func (s *subscriberService) Subscribe(ctx context.Context, req *models.CreateSubscriberRequest) (*models.Subscriber, error) {
	sub, err := s.repo.Save(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("subscriberService.Subscribe: %w", err)
	}
	return sub, nil
}

func (s *subscriberService) Unsubscribe(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("subscriberService.Unsubscribe: %w", err)
	}
	return nil
}
