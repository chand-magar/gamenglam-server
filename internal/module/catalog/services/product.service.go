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
	ErrProductNotFound = errors.New("product not found")
)

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) FindAll(ctx context.Context) ([]models.Product, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("productService.FindAll error: %v", err)
		return nil, fmt.Errorf("productService.FindAll: %w", err)
	}

	return data, nil
}

func (s *productService) FindOne(ctx context.Context, Id uuid.UUID) (*models.Product, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("productService.FindOne error: %v", err)
		return nil, fmt.Errorf("productService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrProductNotFound
	}

	return data, nil
}

func (s *productService) Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	product := &models.Product{
		ProductId:   uuid.New(),
		CategoryNo:  req.CategoryNo,
		SKU:         req.SKU,
		ProductName: req.ProductName,
		Slug:        req.Slug,
		Description: req.Description,
		BasePrice:   req.BasePrice,
		IsFeatured:  req.IsFeatured,
		StockQty:    req.StockQty,
		Status:      models.ProductStatusDraft,
		CreatedBy:   req.CreatedBy,
	}

	created, err := s.repo.Save(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("productService.Create: %w", err)
	}

	return created, nil
}

func (s *productService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateProductRequest) (*models.Product, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("productService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrProductNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("productService.Update: %w", err)
	}

	return updated, nil
}

func (s *productService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("productService.Delete: %w", err)
	}
	if existing == nil {
		return ErrProductNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("productService.Delete: %w", err)
	}

	return nil
}