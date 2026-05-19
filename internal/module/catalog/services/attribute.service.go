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
	ErrAttributeNotFound = errors.New("attribute not found")
)

type attributeService struct {
	repo repository.AttributeRepository
}

func NewAttributeService(repo repository.AttributeRepository) AttributeService {
	return &attributeService{repo: repo}
}

func (s *attributeService) FindAll(ctx context.Context) ([]models.Attribute, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("attributeService.FindAll error: %v", err)
		return nil, fmt.Errorf("attributeService.FindAll: %w", err)
	}

	return data, nil
}

func (s *attributeService) FindOne(ctx context.Context, Id uuid.UUID) (*models.Attribute, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("attributeService.FindOne error: %v", err)
		return nil, fmt.Errorf("attributeService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrAttributeNotFound
	}

	return data, nil
}

func (s *attributeService) Create(ctx context.Context, req *models.CreateAttributeRequest) (*models.Attribute, error) {
	attribute := &models.Attribute{
		AttributeId:   uuid.New(),
		CategoryNo:    req.CategoryNo,
		AttributeName: req.AttributeName,
		Code:          req.Code,
		Type:          req.Type,
		IsRequired:    req.IsRequired,
		IsVariant:     req.IsVariant,
		SortOrder:     req.SortOrder,
		CreatedBy:     req.CreatedBy,
	}

	created, err := s.repo.Save(ctx, attribute)
	if err != nil {
		return nil, fmt.Errorf("attributeService.Create: %w", err)
	}

	return created, nil
}

func (s *attributeService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeRequest) (*models.Attribute, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("attributeService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrAttributeNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("attributeService.Update: %w", err)
	}

	return updated, nil
}

func (s *attributeService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("attributeService.Delete: %w", err)
	}
	if existing == nil {
		return ErrAttributeNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("attributeService.Delete: %w", err)
	}

	return nil
}