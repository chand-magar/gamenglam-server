package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	repository "github.com/chand-magar/gamenglam-server/internal/module/catalog/repository"
)

type reviewService struct {
	repo repository.ReviewRepository
}

func NewReviewService(repo repository.ReviewRepository) ReviewService {
	return &reviewService{repo: repo}
}

func (s *reviewService) GetByProductId(ctx context.Context, productId uuid.UUID) ([]models.Review, error) {
	reviews, err := s.repo.GetByProductId(ctx, productId)
	if err != nil {
		return nil, fmt.Errorf("reviewService.GetByProductId: %w", err)
	}
	if reviews == nil {
		return []models.Review{}, nil
	}
	return reviews, nil
}

func (s *reviewService) FindAll(ctx context.Context, req *models.ReviewListRequest) (*models.ReviewListResponse, error) {
	result, err := s.repo.GetAll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("reviewService.FindAll: %w", err)
	}
	return result, nil
}

func (s *reviewService) Delete(ctx context.Context, reviewId uuid.UUID) error {
	if err := s.repo.Delete(ctx, reviewId); err != nil {
		return fmt.Errorf("reviewService.Delete: %w", err)
	}
	return nil
}

func (s *reviewService) Create(ctx context.Context, productId uuid.UUID, customerNo uint32, req *models.CreateReviewRequest) (*models.Review, error) {
	review := &models.Review{
		ReviewId:   uuid.New(),
		ProductId:  productId,
		CustomerNo: customerNo,
		Rating:     req.Rating,
		Title:      req.Title,
		Body:       req.Body,
	}
	saved, err := s.repo.Save(ctx, review)
	if err != nil {
		return nil, fmt.Errorf("reviewService.Create: %w", err)
	}
	return saved, nil
}
