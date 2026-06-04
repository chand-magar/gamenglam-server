package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	OrderItemNo uint32          `json:"order_item_no"`
	OrderItemId uuid.UUID       `json:"order_item_id"`
	OrderNo     uint32          `json:"order_no"`
	VariantNo   uint32          `json:"variant_no"`
	ProductName string          `json:"product_name"`
	SKU         string          `json:"sku"`
	Attributes  json.RawMessage `json:"attributes"`
	Quantity    int             `json:"quantity"`
	UnitPrice   float64         `json:"unit_price"`
	TotalPrice  float64         `json:"total_price"`
	CreatedAt   time.Time       `json:"created_at"`
}