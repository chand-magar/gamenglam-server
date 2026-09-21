package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
)

type CategoryService interface {
	FindAll(ctx context.Context) ([]models.Category, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.Category, error)
	Create(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateCategoryRequest) (*models.Category, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}
type ProductService interface {
	FindAll(ctx context.Context) ([]models.Product, error)
	FindByCategory(ctx context.Context, categoryId uuid.UUID) ([]models.Product, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.Product, error)
	FindAttributes(ctx context.Context, productId uuid.UUID) ([]models.ProductAttribute, error) // ← add
	Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateProductRequest) (*models.Product, error)
	UpdateImage(ctx context.Context, Id uuid.UUID, imageURL string) (*models.Product, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}
type ProductVariantService interface {
	FindAll(ctx context.Context) ([]models.ProductVariant, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.ProductVariant, error)
	FindOneById(ctx context.Context, variantId uuid.UUID) (*models.EnrichedVariant, error)
	FindByProductId(ctx context.Context, productId uuid.UUID) ([]models.ProductVariant, error)
	Create(ctx context.Context, req *models.CreateProductVariantRequest) (*models.ProductVariant, error)
	BulkCreate(ctx context.Context, req *models.BulkCreateVariantsRequest) (*models.BulkCreateVariantsResponse, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateProductVariantRequest) (*models.ProductVariant, error)
	Delete(ctx context.Context, Id uuid.UUID) error
	FindByProductIdWithAttributes(ctx context.Context, productId uuid.UUID) ([]models.VariantWithAttributes, error)
	GetListing(ctx context.Context, categoryId *uuid.UUID, attributeValues []string) ([]models.VariantListing, error)
}
type AttributeService interface {
	FindAll(ctx context.Context) ([]models.Attribute, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.Attribute, error)
	Create(ctx context.Context, req *models.CreateAttributeRequest) (*models.Attribute, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeRequest) (*models.Attribute, error)
	Delete(ctx context.Context, Id uuid.UUID) error
	FindByCategoryWithValues(ctx context.Context, categoryId uuid.UUID) ([]models.AttributeWithValues, error)
}
type AttributeValueService interface {
	FindAll(ctx context.Context) ([]models.AttributeValue, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.AttributeValue, error)
	FindByAttributeId(ctx context.Context, Id uuid.UUID) ([]models.AttributeValue, error)
	Create(ctx context.Context, req *models.CreateAttributeValueRequest) (*models.AttributeValue, error)
	Update(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeValueRequest) (*models.AttributeValue, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}
type VariantAttributeValueService interface {
	FindAll(ctx context.Context) ([]models.VariantAttributeValue, error)
	FindOne(ctx context.Context, Id uuid.UUID) (*models.VariantAttributeValue, error)
	FindByVariantId(ctx context.Context, variantId uuid.UUID) ([]models.VariantAttributeValue, error)
	Create(ctx context.Context, req *models.CreateVariantAttributeValueRequest) (*models.VariantAttributeValue, error)
	Delete(ctx context.Context, Id uuid.UUID) error
}
