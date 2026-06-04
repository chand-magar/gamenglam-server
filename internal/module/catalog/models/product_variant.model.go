package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	VariantNo     uint32        `json:"variant_no"`
	VariantId     uuid.UUID     `json:"variant_id"`
	ProductNo     uint32        `json:"product_no"`
	SKU           string        `json:"sku"`
	PriceOverride *float64      `json:"price_override"`
	StockQty      int32         `json:"stock_qty"`
	WeightKg      *float64      `json:"weight_kg"`
	ImageURL      *string       `json:"image_url"`
	Status        ProductStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	CreatedBy     *uint32       `json:"created_by"`
	UpdatedAt     time.Time     `json:"updated_at"`
	UpdatedBy     *uint32       `json:"updated_by"`
}

type CreateProductVariantRequest struct {
	ProductNo     uint32   `json:"product_no"     binding:"required"`
	SKU           string   `json:"sku"            binding:"required,min=2,max=100"`
	PriceOverride *float64 `json:"price_override" binding:"omitempty,gt=0"`
	StockQty      int32    `json:"stock_qty"      binding:"min=0"`
	WeightKg      *float64 `json:"weight_kg"      binding:"omitempty,gt=0"`
	ImageURL      *string  `json:"image_url"      binding:"omitempty,max=512"`
	CreatedBy     *uint32  `json:"created_by"`
}

type UpdateProductVariantRequest struct {
	ProductNo     uint32        `json:"product_no"     binding:"required"`
	SKU           string        `json:"sku"            binding:"required,min=2,max=100"`
	PriceOverride *float64      `json:"price_override" binding:"omitempty,gt=0"`
	StockQty      int32         `json:"stock_qty"      binding:"min=0"`
	WeightKg      *float64      `json:"weight_kg"      binding:"omitempty,gt=0"`
	ImageURL      *string       `json:"image_url"      binding:"omitempty,max=512"`
	Status        ProductStatus `json:"status"`
	UpdatedBy     *uint32       `json:"updated_by"`
}

// EnrichedVariant — used only for cross-module provider, not stored
type EnrichedVariant struct {
	VariantNo   uint32
	VariantId   uuid.UUID
	ProductName string
	SKU         string
	Attributes  map[string]interface{}
	Price       float64
	StockQty    int32
}