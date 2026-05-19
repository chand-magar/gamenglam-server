package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ProductNo   uint32        `json:"product_no"`
	ProductId   uuid.UUID     `json:"product_id"`
	CategoryNo  uint32        `json:"category_no"`
	SKU         string        `json:"sku"`
	ProductName string        `json:"product_name"`
	Slug        string        `json:"slug"`
	Description string        `json:"description"`
	BasePrice   float64       `json:"base_price"`
	Status      ProductStatus `json:"status"`
	IsFeatured  bool          `json:"is_featured"`
	StockQty    int32         `json:"stock_qty"`
	CreatedAt   time.Time     `json:"created_at"`
	CreatedBy   *uint32       `json:"created_by"`
	UpdatedAt   time.Time     `json:"updated_at"`
	UpdatedBy   *uint32       `json:"updated_by"`
}

type CreateProductRequest struct {
	CategoryNo  uint32  `json:"category_no"  binding:"required"`
	SKU         string  `json:"sku"          binding:"required,min=2,max=100"`
	ProductName string  `json:"product_name" binding:"required,min=2,max=255"`
	Slug        string  `json:"slug"         binding:"required,min=2,max=255"`
	Description string  `json:"description"`
	BasePrice   float64 `json:"base_price"   binding:"required,gt=0"`
	IsFeatured  bool    `json:"is_featured"`
	StockQty    int32   `json:"stock_qty"    binding:"min=0"`
	CreatedBy   *uint32 `json:"created_by"`
}

type UpdateProductRequest struct {
	CategoryNo  uint32  `json:"category_no"  binding:"required"`
	SKU         string  `json:"sku"          binding:"required,min=2,max=100"`
	ProductName string  `json:"product_name" binding:"required,min=2,max=255"`
	Slug        string  `json:"slug"         binding:"required,min=2,max=255"`
	Description string  `json:"description"`
	BasePrice   float64 `json:"base_price"   binding:"required,gt=0"`
	IsFeatured  bool    `json:"is_featured"`
	StockQty    int32   `json:"stock_qty"    binding:"min=0"`
	Status      ProductStatus `json:"status"`
	UpdatedBy   *uint32 `json:"updated_by"`
}