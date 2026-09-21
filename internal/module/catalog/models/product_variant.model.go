package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	VariantNo       uint32        `json:"-"`
	VariantId       uuid.UUID     `json:"variant_id"`
	ProductNo       uint32        `json:"-"`
	ProductId       *uuid.UUID    `json:"product_id,omitempty"`
	ProductName     *string       `json:"product_name,omitempty"`
	CategoryNo      *uint32       `json:"-"`
	CategoryName    *string       `json:"category_name,omitempty"`
	ProductImageURL *string       `json:"product_image_url,omitempty"`
	SKU             string        `json:"sku"`
	CostPrice       *float64      `json:"cost_price"`    // ← add
	MRP             *float64      `json:"mrp"`           // ← add
	SellingPrice    *float64      `json:"selling_price"` // ← add
	StockQty        int32         `json:"stock_qty"`
	WeightKg        *float64      `json:"weight_kg"`
	ImageURL        *string       `json:"image_url"`
	Status          ProductStatus `json:"status"`
	BasePrice       *float64      `json:"base_price,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	CreatedBy       *uint32       `json:"-"`
	UpdatedAt       time.Time     `json:"updated_at"`
	UpdatedBy       *uint32       `json:"-"`
}

type CreateProductVariantRequest struct {
	ProductNo uint32   `json:"product_no"     binding:"required"`
	SKU       string   `json:"sku"            binding:"required,min=2,max=100"`
	StockQty  int32    `json:"stock_qty"      binding:"min=0"`
	WeightKg  *float64 `json:"weight_kg"      binding:"omitempty,gt=0"`
	ImageURL  *string  `json:"image_url"      binding:"omitempty,max=512"`
	CreatedBy *uint32  `json:"created_by"`
}

type UpdateProductVariantRequest struct {
	ProductNo    uint32        `json:"product_no"     binding:"omitempty"`
	SKU          string        `json:"sku"            binding:"required,min=2,max=100"`
	CostPrice    *float64      `json:"cost_price"     binding:"omitempty,gt=0"` // ← add
	MRP          *float64      `json:"mrp"            binding:"omitempty,gt=0"` // ← add
	SellingPrice *float64      `json:"selling_price"  binding:"omitempty,gt=0"` // ← add
	StockQty     int32         `json:"stock_qty"      binding:"min=0"`
	WeightKg     *float64      `json:"weight_kg"      binding:"omitempty,gt=0"`
	ImageURL     *string       `json:"image_url"      binding:"omitempty,max=512"`
	Status       ProductStatus `json:"status"         binding:"omitempty"`
	UpdatedBy    *uint32       `json:"-"`
}

// ── Bulk create ───────────────────────────────────────────────────────────────

// BulkVariantItem — one generated variant with its attribute value assignments
type BulkVariantItem struct {
	SKU               string   `json:"sku"                  binding:"required,min=2,max=100"`
	CostPrice         *float64 `json:"cost_price"           binding:"omitempty,gt=0"`
	MRP               *float64 `json:"mrp"                  binding:"omitempty,gt=0"`
	SellingPrice      *float64 `json:"selling_price"        binding:"omitempty,gt=0"`
	StockQty          int32    `json:"stock_qty"            binding:"min=0"`
	WeightKg          *float64 `json:"weight_kg"            binding:"omitempty,gt=0"`
	ImageURL          *string  `json:"image_url"            binding:"omitempty,max=512"`
	AttributeValueIds []string `json:"attribute_value_ids"  binding:"required,min=1"` // ← UUIDs
}

type BulkCreateVariantsRequest struct {
	ProductId uuid.UUID         `json:"product_id" binding:"required"` // ← UUID
	Variants  []BulkVariantItem `json:"variants"   binding:"required,min=1"`
}

type BulkCreateVariantsResponse struct {
	Created  int              `json:"created"`
	Variants []ProductVariant `json:"variants"`
}

// EnrichedVariant — cross-module provider only
type EnrichedVariant struct {
	VariantNo   uint32
	VariantId   uuid.UUID
	ProductName string
	SKU         string
	Attributes  map[string]interface{}
	Price       float64
	StockQty    int32
}
type VariantWithAttributes struct {
	VariantId  uuid.UUID         `json:"variant_id"`
	SKU        string            `json:"sku"`
	StockQty   int32             `json:"stock_qty"`
	WeightKg   *float64          `json:"weight_kg"`
	Status     ProductStatus     `json:"status"`
	Attributes map[string]string `json:"attributes"` // code → value
}
type VariantListing struct {
	VariantId    uuid.UUID         `json:"variant_id"`
	ProductId    uuid.UUID         `json:"product_id"`
	ProductName  string            `json:"product_name"`
	Slug         string            `json:"slug"`
	SKU          string            `json:"sku"`
	CostPrice    *float64          `json:"cost_price"`
	MRP          *float64          `json:"mrp"`
	SellingPrice *float64          `json:"selling_price"`
	StockQty     int32             `json:"stock_qty"`
	ImageURL     *string           `json:"image_url"`
	CategoryId   uuid.UUID         `json:"category_id"`
	IsFeatured   bool              `json:"is_featured"`
	Attributes   map[string]string `json:"attributes"` // code → value
}
