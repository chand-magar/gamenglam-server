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
	ErrCategoryNotFound = errors.New("category not found")
)

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) FindAll(ctx context.Context) ([]models.Category, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("categoryService.FindAll error: %v", err)
		return nil, fmt.Errorf("categoryService.FindAll: %w", err)
	}

	return data, nil
}

func (s *categoryService) FindOne(ctx context.Context, Id uuid.UUID) (*models.Category, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("categoryService.FindOne error: %v", err)
		return nil, fmt.Errorf("categoryService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrCategoryNotFound
	}

	return data, nil
}

func (s *categoryService) Create(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	category := &models.Category{
		CategoryId:   uuid.New(),
		ParentId:     req.ParentId, // ← UUID
		CategoryName: req.CategoryName,
		Slug:         req.Slug,
		Details:      req.Details,
		Status:       models.StatusActive,
		CreatedBy:    req.CreatedBy,
	}
	created, err := s.repo.Save(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("categoryService.Create: %w", err)
	}
	return created, nil
}

func (s *categoryService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateCategoryRequest) (*models.Category, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("categoryService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrCategoryNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("categoryService.Update: %w", err)
	}

	return updated, nil
}

func (s *categoryService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("categoryService.Delete: %w", err)
	}
	if existing == nil {
		return ErrCategoryNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("categoryService.Delete: %w", err)
	}

	return nil
}
