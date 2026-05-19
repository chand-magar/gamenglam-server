package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
)

type CategoryRepository interface {
	GetAll(ctx context.Context) ([]models.Category, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Category, error)
	Save(ctx context.Context, category *models.Category) (*models.Category, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateCategoryRequest) (*models.Category, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type ProductRepository interface {
	GetAll(ctx context.Context) ([]models.Product, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Product, error)
	Save(ctx context.Context, product *models.Product) (*models.Product, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateProductRequest) (*models.Product, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type ProductVariantRepository interface {
	GetAll(ctx context.Context) ([]models.ProductVariant, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.ProductVariant, error)
	Save(ctx context.Context, product_variant *models.ProductVariant) (*models.ProductVariant, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateProductVariantRequest) (*models.ProductVariant, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type AttributeRepository interface {
	GetAll(ctx context.Context) ([]models.Attribute, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Attribute, error)
	Save(ctx context.Context, attribute *models.Attribute) (*models.Attribute, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeRequest) (*models.Attribute, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type AttributeValueRepository interface {
	GetAll(ctx context.Context) ([]models.AttributeValue, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.AttributeValue, error)
	Save(ctx context.Context, attributeValue *models.AttributeValue) (*models.AttributeValue, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeValueRequest) (*models.AttributeValue, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type VariantAttributeValueRepository interface {
	GetAll(ctx context.Context) ([]models.VariantAttributeValue, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.VariantAttributeValue, error)
	Save(ctx context.Context, variantAttributeValue *models.VariantAttributeValue) (*models.VariantAttributeValue, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}