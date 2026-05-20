package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/repository"
)

var (
	ErrVariantAttributeValueNotFound = errors.New("variant attribute value not found")
)

type variantAttributeValueService struct {
	repo repository.VariantAttributeValueRepository
}

func NewVariantAttributeValueService(repo repository.VariantAttributeValueRepository) VariantAttributeValueService {
	return &variantAttributeValueService{repo: repo}
}

func (s *variantAttributeValueService) FindAll(ctx context.Context) ([]models.VariantAttributeValue, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("variantAttributeValueService.FindAll error: %v", err)
		return nil, fmt.Errorf("variantAttributeValueService.FindAll: %w", err)
	}

	return data, nil
}

func (s *variantAttributeValueService) FindOne(ctx context.Context, Id uuid.UUID) (*models.VariantAttributeValue, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("variantAttributeValueService.FindOne error: %v", err)
		return nil, fmt.Errorf("variantAttributeValueService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrVariantAttributeValueNotFound
	}

	return data, nil
}

func (s *variantAttributeValueService) Create(ctx context.Context, req *models.CreateVariantAttributeValueRequest) (*models.VariantAttributeValue, error) {
	variantAttributeValue := &models.VariantAttributeValue{
		VariantAttributeValueId: uuid.New(),
		VariantNo:               req.VariantNo,
		AttributeValueNo:        req.AttributeValueNo,
		CreatedBy:               req.CreatedBy,
	}

	created, err := s.repo.Save(ctx, variantAttributeValue)
	if err != nil {
		return nil, fmt.Errorf("variantAttributeValueService.Create: %w", err)
	}

	return created, nil
}

func (s *variantAttributeValueService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("variantAttributeValueService.Delete: %w", err)
	}
	if existing == nil {
		return ErrVariantAttributeValueNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("variantAttributeValueService.Delete: %w", err)
	}

	return nil
}