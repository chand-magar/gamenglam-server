package models

import (
	"time"

	"github.com/google/uuid"
)

type Attribute struct {
	AttributeNo   uint32        `json:"-"`
	AttributeId   uuid.UUID     `json:"attribute_id"`
	CategoryNo    uint32        `json:"-"`
	CategoryId    uuid.UUID     `json:"category_id"`
	AttributeName string        `json:"attribute_name"`
	Code          string        `json:"code"`
	Type          AttributeType `json:"type"`
	IsRequired    bool          `json:"is_required"`
	IsVariant     bool          `json:"is_variant"`
	ShowInFilter  bool          `json:"show_in_filter"` // ← add
	SortOrder     int32         `json:"sort_order"`
	Status        StatusEnum    `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	CreatedBy     *uint32       `json:"-"`
	UpdatedAt     time.Time     `json:"updated_at"`
	UpdatedBy     *uint32       `json:"-"`
}
type CreateAttributeRequest struct {
	CategoryId    uuid.UUID     `json:"category_id"`
	AttributeName string        `json:"attribute_name" binding:"required,min=2,max=100"`
	Code          string        `json:"code"           binding:"required,min=2,max=50"`
	Type          AttributeType `json:"type"           binding:"required"`
	IsRequired    bool          `json:"is_required"`
	IsVariant     bool          `json:"is_variant"`
	ShowInFilter  bool          `json:"show_in_filter"` // ← add
	SortOrder     int32         `json:"sort_order"`
	CreatedBy     *uint32       `json:"-"`
}
type UpdateAttributeRequest struct {
	CategoryId    uuid.UUID     `json:"category_id"`
	AttributeName string        `json:"attribute_name" binding:"required,min=2,max=100"`
	Code          string        `json:"code"           binding:"required,min=2,max=50"`
	Type          AttributeType `json:"type"           binding:"required"`
	IsRequired    bool          `json:"is_required"`
	IsVariant     bool          `json:"is_variant"`
	ShowInFilter  bool          `json:"show_in_filter"` // ← add
	SortOrder     int32         `json:"sort_order"`
	UpdatedBy     *uint32       `json:"-"`
}
type AttributeValueOption struct {
	AttributeValueId string `json:"attribute_value_id"`
	DisplayLabel     string `json:"display_label"`
	Value            string `json:"value"`
	ColorHex         string `json:"color_hex,omitempty"`
}
type ProductAttribute struct {
	AttributeId   string                 `json:"attribute_id"`
	AttributeName string                 `json:"attribute_name"`
	Code          string                 `json:"code"`
	Type          string                 `json:"type"`
	Values        []AttributeValueOption `json:"values"`
}
type AttributeWithValues struct {
	AttributeId   string                 `json:"attribute_id"`
	AttributeName string                 `json:"attribute_name"`
	Code          string                 `json:"code"`
	Type          string                 `json:"type"`
	Values        []AttributeValueOption `json:"values"`
}
