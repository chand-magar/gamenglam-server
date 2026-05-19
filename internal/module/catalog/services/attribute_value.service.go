package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catelog/models"
	"github.com/chand-magar/gamenglam-server/internal/module/catelog/repository"
)

var (
	ErrAttributeValueNotFound = errors.New("attribute value not found")
)

type attributeValueService struct {
	repo repository.AttributeValueRepository
}

func NewAttributeValueService(repo repository.AttributeValueRepository) AttributeValueService {
	return &attributeValueService{repo: repo}
}

func (s *attributeValueService) FindAll(ctx context.Context) ([]models.AttributeValue, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("attributeValueService.FindAll error: %v", err)
		return nil, fmt.Errorf("attributeValueService.FindAll: %w", err)
	}

	return data, nil
}

func (s *attributeValueService) FindOne(ctx context.Context, Id uuid.UUID) (*models.AttributeValue, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("attributeValueService.FindOne error: %v", err)
		return nil, fmt.Errorf("attributeValueService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrAttributeValueNotFound
	}

	return data, nil
}

func (s *attributeValueService) Create(ctx context.Context, req *models.CreateAttributeValueRequest) (*models.AttributeValue, error) {
	attributeValue := &models.AttributeValue{
		AttributeValueId: uuid.New(),
		AttributeNo:      req.AttributeNo,
		Value:            req.Value,
		DisplayLabel:     req.DisplayLabel,
		ColorHex:         req.ColorHex,
		SortOrder:        req.SortOrder,
		CreatedBy:        req.CreatedBy,
	}

	created, err := s.repo.Save(ctx, attributeValue)
	if err != nil {
		return nil, fmt.Errorf("attributeValueService.Create: %w", err)
	}

	return created, nil
}

func (s *attributeValueService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeValueRequest) (*models.AttributeValue, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("attributeValueService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrAttributeValueNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("attributeValueService.Update: %w", err)
	}

	return updated, nil
}

func (s *attributeValueService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("attributeValueService.Delete: %w", err)
	}
	if existing == nil {
		return ErrAttributeValueNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("attributeValueService.Delete: %w", err)
	}

	return nil
}