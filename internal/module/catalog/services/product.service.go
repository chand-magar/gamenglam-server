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
		ProductId:      uuid.New(), // ← generate here
		CategoryId:     &req.CategoryId,
		SKU:            req.SKU,
		ProductName:    req.ProductName,
		Slug:           req.Slug,
		Description:    req.Description,
		ImageURL:       req.ImageURL,
		Features:       req.Features,
		Specifications: req.Specifications,
		IsFeatured:     req.IsFeatured,
		CreatedBy:      req.CreatedBy,
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

func (s *productService) UpdateImage(ctx context.Context, Id uuid.UUID, imageURL string) (*models.Product, error) {
	updated, err := s.repo.ModifyImage(ctx, Id, imageURL)
	if err != nil {
		return nil, fmt.Errorf("productService.UpdateImage: %w", err)
	}
	return updated, nil
}

func (s *productService) FindByCategory(ctx context.Context, categoryId uuid.UUID) ([]models.Product, error) {
	products, err := s.repo.GetByCategory(ctx, categoryId)
	if err != nil {
		return nil, fmt.Errorf("productService.FindByCategory: %w", err)
	}
	return products, nil
}

func (s *productService) FindAttributes(ctx context.Context, productId uuid.UUID) ([]models.ProductAttribute, error) {
	attrs, err := s.repo.GetAttributesByProductId(ctx, productId)
	if err != nil {
		return nil, fmt.Errorf("productService.FindAttributes: %w", err)
	}
	return attrs, nil
}
