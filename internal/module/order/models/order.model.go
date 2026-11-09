package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderNo      uint32      `json:"-"`
	OrderId      uuid.UUID   `json:"order_id"`
	CustomerNo   uint32      `json:"-"`
	AddressNo    uint32      `json:"-"`
	CustomerName string      `json:"customer_name,omitempty"`
	ItemCount    int         `json:"item_count,omitempty"`
	OrderRef     string      `json:"order_ref"`
	Subtotal     float64     `json:"subtotal"`
	Discount     float64     `json:"discount"`
	Tax          float64     `json:"tax"`
	DeliveryFee  float64     `json:"delivery_fee"`
	Total        float64     `json:"total"`
	OrderStatus  OrderStatus `json:"order_status"`
	Notes        *string     `json:"notes"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Items        []OrderItem `json:"items"`
}

type OrderListRequest struct {
	Page   int
	Limit  int
	Status string
	Search string
}

type OrderListResponse struct {
	Orders     []Order `json:"orders"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
}

type OrderStatusHistory struct {
	HistoryNo  uint32       `json:"-"` // ← hide
	HistoryId  uuid.UUID    `json:"history_id"`
	OrderNo    uint32       `json:"-"` // ← hide
	FromStatus *OrderStatus `json:"from_status"`
	ToStatus   OrderStatus  `json:"to_status"`
	ChangedBy  *uint32      `json:"-"` // ← hide
	Note       *string      `json:"note"`
	ChangedAt  time.Time    `json:"changed_at"`
}
type PlaceOrderRequest struct {
	CustomerNo  uint32                  `json:"customer_no"  binding:"required"`
	AddressId   uuid.UUID               `json:"address_id"   binding:"required"` // ← UUID
	Items       []PlaceOrderItemRequest `json:"items"        binding:"required,min=1"`
	DeliveryFee float64                 `json:"delivery_fee"`
	Discount    float64                 `json:"discount"`
	Notes       *string                 `json:"notes"`
}
type PlaceOrderItemRequest struct {
	VariantId uuid.UUID `json:"variant_id" binding:"required"` // ← UUID
	Quantity  int32     `json:"quantity"   binding:"required,gt=0"`
	UnitPrice float64   `json:"unit_price"`
}
type UpdateOrderStatusRequest struct {
	ToStatus  OrderStatus `json:"to_status"  binding:"required"`
	Note      *string     `json:"note"`
	ChangedBy *uint32     `json:"changed_by"`
}

// Keep AddressNo for internal use only — resolved from AddressId
type resolvedOrder struct {
	AddressNo uint32
}
