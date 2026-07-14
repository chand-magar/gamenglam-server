package models

import (
	"time"

	"github.com/google/uuid"
)

type Attribute struct {
	AttributeNo   uint32        `json:"attribute_no"`
	AttributeId   uuid.UUID     `json:"attribute_id"`
	CategoryNo    uint32        `json:"category_no"`
	AttributeName string        `json:"attribute_name"`
	Code          string        `json:"code"`
	Type          AttributeType `json:"type"`
	IsRequired    bool          `json:"is_required"`
	IsVariant     bool          `json:"is_variant"`
	SortOrder     int32         `json:"sort_order"`
	Status        StatusEnum    `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	CreatedBy     *uint32       `json:"created_by"`
	UpdatedAt     time.Time     `json:"updated_at"`
	UpdatedBy     *uint32       `json:"updated_by"`
}

type CreateAttributeRequest struct {
	CategoryNo    uint32        `json:"category_no"    binding:"required"`
	AttributeName string        `json:"attribute_name" binding:"required,min=2,max=100"`
	Code          string        `json:"code"           binding:"required,min=2,max=50"`
	Type          AttributeType `json:"type"           binding:"required"`
	IsRequired    bool          `json:"is_required"`
	IsVariant     bool          `json:"is_variant"`
	SortOrder     int32         `json:"sort_order"`
	CreatedBy     *uint32       `json:"created_by"`
}

type UpdateAttributeRequest struct {
	CategoryNo    uint32        `json:"category_no"    binding:"required"`
	AttributeName string        `json:"attribute_name" binding:"required,min=2,max=100"`
	Code          string        `json:"code"           binding:"required,min=2,max=50"`
	Type          AttributeType `json:"type"           binding:"required"`
	IsRequired    bool          `json:"is_required"`
	IsVariant     bool          `json:"is_variant"`
	SortOrder     int32         `json:"sort_order"`
	UpdatedBy     *uint32       `json:"updated_by"`
}
