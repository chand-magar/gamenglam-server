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
	ErrProductVariantNotFound = errors.New("product variant not found")
)

type productVariantService struct {
	repo repository.ProductVariantRepository
}

func NewProductVariantService(repo repository.ProductVariantRepository) ProductVariantService {
	return &productVariantService{repo: repo}
}

func (s *productVariantService) FindAll(ctx context.Context) ([]models.ProductVariant, error) {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Printf("productVariantService.FindAll error: %v", err)
		return nil, fmt.Errorf("productVariantService.FindAll: %w", err)
	}

	return data, nil
}

func (s *productVariantService) FindOne(ctx context.Context, Id uuid.UUID) (*models.ProductVariant, error) {
	data, err := s.repo.GetById(ctx, Id)
	if err != nil {
		log.Printf("productVariantService.FindOne error: %v", err)
		return nil, fmt.Errorf("productVariantService.FindOne: %w", err)
	}
	if data == nil {
		return nil, ErrProductVariantNotFound
	}

	return data, nil
}

func (s *productVariantService) FindOneByNo(ctx context.Context, variantNo uint32) (*models.EnrichedVariant, error) {
	variant, err := s.repo.FindOneByNo(ctx, variantNo)
	if err != nil {
		log.Printf("productVariantService.FindOneByNo error: %v", err)
		return nil, fmt.Errorf("productVariantService.FindOneByNo: %w", err)
	}
	return variant, nil
}

func (s *productVariantService) Create(ctx context.Context, req *models.CreateProductVariantRequest) (*models.ProductVariant, error) {
	productVariant := &models.ProductVariant{
		VariantId:     uuid.New(),
		ProductNo:     req.ProductNo,
		SKU:           req.SKU,
		PriceOverride: req.PriceOverride,
		StockQty:      req.StockQty,
		WeightKg:      req.WeightKg,
		ImageURL:      req.ImageURL,
		Status:        models.ProductStatusDraft,
		CreatedBy:     req.CreatedBy,
	}

	created, err := s.repo.Save(ctx, productVariant)
	if err != nil {
		return nil, fmt.Errorf("productVariantService.Create: %w", err)
	}

	return created, nil
}

func (s *productVariantService) Update(ctx context.Context, Id uuid.UUID, req *models.UpdateProductVariantRequest) (*models.ProductVariant, error) {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("productVariantService.Update: %w", err)
	}
	if existing == nil {
		return nil, ErrProductVariantNotFound
	}

	updated, err := s.repo.Modify(ctx, Id, req)
	if err != nil {
		return nil, fmt.Errorf("productVariantService.Update: %w", err)
	}

	return updated, nil
}

func (s *productVariantService) Delete(ctx context.Context, Id uuid.UUID) error {
	existing, err := s.repo.GetById(ctx, Id)
	if err != nil {
		return fmt.Errorf("productVariantService.Delete: %w", err)
	}
	if existing == nil {
		return ErrProductVariantNotFound
	}

	if err := s.repo.Remove(ctx, Id); err != nil {
		return fmt.Errorf("productVariantService.Delete: %w", err)
	}

	return nil
}