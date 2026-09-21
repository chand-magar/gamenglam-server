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
	repo    repository.ProductVariantRepository
	vavRepo repository.VariantAttributeValueRepository // ← add
}

func NewProductVariantService(
	repo repository.ProductVariantRepository,
	vavRepo repository.VariantAttributeValueRepository, // ← add
) ProductVariantService {
	return &productVariantService{repo: repo, vavRepo: vavRepo}
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

func (s *productVariantService) FindOneById(ctx context.Context, variantId uuid.UUID) (*models.EnrichedVariant, error) {
	variant, err := s.repo.FindOneById(ctx, variantId)
	if err != nil {
		log.Printf("productVariantService.FindOneById error: %v", err)
		return nil, fmt.Errorf("productVariantService.FindOneById: %w", err)
	}
	return variant, nil
}

func (s *productVariantService) Create(ctx context.Context, req *models.CreateProductVariantRequest) (*models.ProductVariant, error) {
	productVariant := &models.ProductVariant{
		VariantId: uuid.New(),
		ProductNo: req.ProductNo,
		SKU:       req.SKU,
		StockQty:  req.StockQty,
		WeightKg:  req.WeightKg,
		ImageURL:  req.ImageURL,
		Status:    models.ProductStatusDraft,
		CreatedBy: req.CreatedBy,
	}

	created, err := s.repo.Save(ctx, productVariant)
	if err != nil {
		return nil, fmt.Errorf("productVariantService.Create: %w", err)
	}
	return created, nil
}

func (s *productVariantService) BulkCreate(
	ctx context.Context,
	req *models.BulkCreateVariantsRequest,
) (*models.BulkCreateVariantsResponse, error) {

	// Resolve product_no from product_id
	productNo, err := s.repo.GetProductNoByProductId(ctx, req.ProductId)
	if err != nil {
		return nil, fmt.Errorf("productVariantService.BulkCreate: resolve product: %w", err)
	}

	var created []models.ProductVariant
	for _, item := range req.Variants {
		variant := &models.ProductVariant{
			VariantId:    uuid.New(),
			ProductNo:    productNo,
			SKU:          item.SKU,
			CostPrice:    item.CostPrice,
			MRP:          item.MRP,
			SellingPrice: item.SellingPrice,
			StockQty:     item.StockQty,
			WeightKg:     item.WeightKg,
			ImageURL:     item.ImageURL,
			Status:       models.ProductStatusDraft,
		}

		saved, err := s.repo.Save(ctx, variant)
		if err != nil {
			log.Printf("productVariantService.BulkCreate: save variant: %v", err)
			return nil, fmt.Errorf("productVariantService.BulkCreate: save variant %s: %w", item.SKU, err)
		}

		// Resolve attribute_value_no from attribute_value_id for each value
		for _, valueId := range item.AttributeValueIds {
			avId, parseErr := uuid.Parse(valueId)
			if parseErr != nil {
				return nil, fmt.Errorf("productVariantService.BulkCreate: parse attribute_value_id: %w", parseErr)
			}

			valueNo, err := s.vavRepo.GetAttributeValueNoById(ctx, avId)
			if err != nil {
				return nil, fmt.Errorf("productVariantService.BulkCreate: resolve attribute value: %w", err)
			}

			vav := &models.VariantAttributeValue{
				VariantAttributeValueId: uuid.New(),
				VariantNo:               saved.VariantNo,
				AttributeValueNo:        valueNo,
			}
			if _, err := s.vavRepo.Save(ctx, vav); err != nil {
				return nil, fmt.Errorf("productVariantService.BulkCreate: save vav: %w", err)
			}
		}

		created = append(created, *saved)
	}

	return &models.BulkCreateVariantsResponse{
		Created:  len(created),
		Variants: created,
	}, nil
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

func (s *productVariantService) FindByProductId(ctx context.Context, productId uuid.UUID) ([]models.ProductVariant, error) {
	data, err := s.repo.GetByProductId(ctx, productId)
	if err != nil {
		log.Printf("productVariantService.FindByProductId error: %v", err)
		return nil, fmt.Errorf("productVariantService.FindByProductId: %w", err)
	}
	return data, nil
}

func (s *productVariantService) FindByProductIdWithAttributes(ctx context.Context, productId uuid.UUID) ([]models.VariantWithAttributes, error) {
	variants, err := s.repo.GetByProductIdWithAttributes(ctx, productId)
	if err != nil {
		return nil, fmt.Errorf("productVariantService.FindByProductIdWithAttributes: %w", err)
	}
	return variants, nil
}

func (s *productVariantService) GetListing(ctx context.Context, categoryId *uuid.UUID, attributeValues []string) ([]models.VariantListing, error) {
	return s.repo.GetListing(ctx, categoryId, attributeValues)
}
