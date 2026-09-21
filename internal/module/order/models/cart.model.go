package models

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	CartNo     uint32     `json:"cart_no"`
	CartId     uuid.UUID  `json:"cart_id"`
	CustomerNo uint32     `json:"customer_no"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Items      []CartItem `json:"items"`
}

type CartItem struct {
	CartItemNo uint32    `json:"cart_item_no"`
	CartItemId uuid.UUID `json:"cart_item_id"`
	CartNo     uint32    `json:"cart_no"`
	VariantNo  uint32    `json:"variant_no"`
	VariantId  uuid.UUID `json:"variant_id"`
	Quantity   int32     `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AddCartItemRequest struct {
	VariantId uuid.UUID `json:"variant_id" binding:"required"` // ← UUID
	Quantity  int       `json:"quantity"   binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}
