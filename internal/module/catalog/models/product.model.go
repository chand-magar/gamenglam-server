package models

import (
	"time"

	"github.com/google/uuid"
)

type SpecItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Product struct {
	ProductNo      uint32        `json:"-"`
	ProductId      uuid.UUID     `json:"product_id"`
	CategoryNo     uint32        `json:"-"`
	CategoryId     *uuid.UUID    `json:"category_id,omitempty"`
	SKU            string        `json:"sku"`
	ProductName    string        `json:"product_name"`
	Slug           string        `json:"slug"`
	Description    *string       `json:"description"`
	ImageURL       *string       `json:"image_url"`
	Features       []string      `json:"features"`
	Specifications []SpecItem    `json:"specifications"`
	Status         ProductStatus `json:"status"`
	IsFeatured     bool          `json:"is_featured"`
	CreatedAt      time.Time     `json:"created_at"`
	CreatedBy      *uint32       `json:"-"`
	UpdatedAt      time.Time     `json:"updated_at"`
	UpdatedBy      *uint32       `json:"-"`
	// Default variant fields — populated from cheapest active variant
	DefaultVariantId    *uuid.UUID `json:"default_variant_id,omitempty"`
	DefaultSellingPrice *float64   `json:"default_selling_price,omitempty"`
	DefaultMRP          *float64   `json:"default_mrp,omitempty"`
	DefaultStockQty     *int32     `json:"default_stock_qty,omitempty"`
}
type CreateProductRequest struct {
	ProductId      uuid.UUID  `json:"-"`
	CategoryId     uuid.UUID  `json:"category_id"  binding:"required"` // ← UUID
	SKU            string     `json:"sku"           binding:"required,min=2,max=100"`
	ProductName    string     `json:"product_name"  binding:"required,min=2,max=255"`
	Slug           string     `json:"slug"          binding:"required,min=2,max=255"`
	Description    *string    `json:"description"   binding:"omitempty"`
	ImageURL       *string    `json:"image_url"     binding:"omitempty,max=512"`
	Features       []string   `json:"features"`
	Specifications []SpecItem `json:"specifications"`
	IsFeatured     bool       `json:"is_featured"`
	CreatedBy      *uint32    `json:"-"`
}

type UpdateProductRequest struct {
	CategoryId     uuid.UUID     `json:"category_id"   binding:"required"` // ← UUID
	SKU            string        `json:"sku"            binding:"required,min=2,max=100"`
	ProductName    string        `json:"product_name"   binding:"required,min=2,max=255"`
	Slug           string        `json:"slug"           binding:"required,min=2,max=255"`
	Description    *string       `json:"description"    binding:"omitempty"`
	ImageURL       *string       `json:"image_url"      binding:"omitempty,max=512"`
	Features       []string      `json:"features"`
	Specifications []SpecItem    `json:"specifications"`
	Status         ProductStatus `json:"status"         binding:"omitempty"`
	IsFeatured     bool          `json:"is_featured"`
	UpdatedBy      *uint32       `json:"-"`
}
