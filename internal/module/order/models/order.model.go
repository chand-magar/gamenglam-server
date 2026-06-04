package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderNo     uint32      `json:"order_no"`
	OrderId     uuid.UUID   `json:"order_id"`
	CustomerNo  uint32      `json:"customer_no"`
	AddressNo   uint32      `json:"address_no"`
	OrderRef    string      `json:"order_ref"`
	Subtotal    float64     `json:"subtotal"`
	Discount    float64     `json:"discount"`
	Tax         float64     `json:"tax"`
	DeliveryFee float64     `json:"delivery_fee"`
	Total       float64     `json:"total"`
	OrderStatus OrderStatus `json:"order_status"`
	Notes       *string     `json:"notes"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Items       []OrderItem `json:"items"`
}

type OrderStatusHistory struct {
	HistoryNo  uint32      `json:"history_no"`
	HistoryId  uuid.UUID   `json:"history_id"`
	OrderNo    uint32      `json:"order_no"`
	FromStatus *OrderStatus `json:"from_status"`
	ToStatus   OrderStatus `json:"to_status"`
	ChangedBy  *uint32     `json:"changed_by"`
	Note       *string     `json:"note"`
	ChangedAt  time.Time   `json:"changed_at"`
}

type PlaceOrderRequest struct {
	CustomerNo  uint32               `json:"customer_no"  binding:"required"`
	AddressNo   uint32               `json:"address_no"   binding:"required"`
	Items       []PlaceOrderItemRequest `json:"items"     binding:"required,min=1"`
	DeliveryFee float64              `json:"delivery_fee"`
	Discount    float64              `json:"discount"`
	Notes       *string              `json:"notes"`
}

type PlaceOrderItemRequest struct {
	VariantNo uint32 `json:"variant_no" binding:"required"`
	Quantity  int    `json:"quantity"   binding:"required,min=1"`
}

type UpdateOrderStatusRequest struct {
	ToStatus  OrderStatus `json:"to_status"  binding:"required"`
	Note      *string     `json:"note"`
	ChangedBy *uint32     `json:"changed_by"`
}