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
	GetByCategory(ctx context.Context, categoryId uuid.UUID) ([]models.Product, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Product, error)
	GetAttributesByProductId(ctx context.Context, productId uuid.UUID) ([]models.ProductAttribute, error)
	Save(ctx context.Context, product *models.Product) (*models.Product, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateProductRequest) (*models.Product, error)
	ModifyImage(ctx context.Context, Id uuid.UUID, imageURL string) (*models.Product, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type ProductVariantRepository interface {
	GetAll(ctx context.Context) ([]models.ProductVariant, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.ProductVariant, error)
	FindOneById(ctx context.Context, variantId uuid.UUID) (*models.EnrichedVariant, error)
	GetByProductId(ctx context.Context, productId uuid.UUID) ([]models.ProductVariant, error)
	Save(ctx context.Context, product_variant *models.ProductVariant) (*models.ProductVariant, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateProductVariantRequest) (*models.ProductVariant, error)
	Remove(ctx context.Context, Id uuid.UUID) error
	GetByProductIdWithAttributes(ctx context.Context, productId uuid.UUID) ([]models.VariantWithAttributes, error)
	GetProductNoByProductId(ctx context.Context, productId uuid.UUID) (uint32, error)
	GetListing(ctx context.Context, categoryId *uuid.UUID, attributeValues []string) ([]models.VariantListing, error)
}

type AttributeRepository interface {
	GetAll(ctx context.Context) ([]models.Attribute, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.Attribute, error)
	Save(ctx context.Context, attribute *models.Attribute) (*models.Attribute, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeRequest) (*models.Attribute, error)
	Remove(ctx context.Context, Id uuid.UUID) error
	GetByCategoryWithValues(ctx context.Context, categoryId uuid.UUID) ([]models.AttributeWithValues, error)
}

type AttributeValueRepository interface {
	GetAll(ctx context.Context) ([]models.AttributeValue, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.AttributeValue, error)
	GetByAttributeId(ctx context.Context, attributeId uuid.UUID) ([]models.AttributeValue, error)
	Save(ctx context.Context, attributeValue *models.AttributeValue) (*models.AttributeValue, error)
	Modify(ctx context.Context, Id uuid.UUID, req *models.UpdateAttributeValueRequest) (*models.AttributeValue, error)
	Remove(ctx context.Context, Id uuid.UUID) error
}

type VariantAttributeValueRepository interface {
	GetAll(ctx context.Context) ([]models.VariantAttributeValue, error)
	GetById(ctx context.Context, Id uuid.UUID) (*models.VariantAttributeValue, error)
	GetByVariantId(ctx context.Context, variantId uuid.UUID) ([]models.VariantAttributeValue, error)
	Save(ctx context.Context, variantAttributeValue *models.VariantAttributeValue) (*models.VariantAttributeValue, error)
	Remove(ctx context.Context, Id uuid.UUID) error
	GetAttributeValueNoById(ctx context.Context, attributeValueId uuid.UUID) (uint32, error)
}
type VariantImageRepository interface {
	GetByVariantId(ctx context.Context, variantId uuid.UUID) ([]models.ProductVariantImage, error)
	GetById(ctx context.Context, imageId uuid.UUID) (*models.ProductVariantImage, error)
	GetVariantNoById(ctx context.Context, variantId uuid.UUID) (uint32, error)
	GetPrimaryByVariantIds(ctx context.Context, variantIds []uuid.UUID) (map[string]string, error) // ← add
	Save(ctx context.Context, img *models.ProductVariantImage) (*models.ProductVariantImage, error)
	BulkSave(ctx context.Context, variantNo uint32, urls []string) ([]models.ProductVariantImage, error)
	Remove(ctx context.Context, imageId uuid.UUID) error
	RemoveAllByVariantId(ctx context.Context, variantId uuid.UUID) error
}
