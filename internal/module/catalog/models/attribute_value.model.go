package models

import (
	"time"

	"github.com/google/uuid"
)

type AttributeValue struct {
	AttributeValueNo uint32    `json:"attribute_value_no"`
	AttributeValueId uuid.UUID `json:"attribute_value_id"`
	AttributeNo      uint32    `json:"attribute_no"`
	Value            string    `json:"value"`
	DisplayLabel     string    `json:"display_label"`
	ColorHex         *string   `json:"color_hex"`
	SortOrder        int32     `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        *uint32   `json:"created_by"`
	UpdatedAt        time.Time `json:"updated_at"`
	UpdatedBy        *uint32   `json:"updated_by"`
}

type CreateAttributeValueRequest struct {
	AttributeNo  uint32  `json:"attribute_no"  binding:"required"`
	Value        string  `json:"value"         binding:"required,min=1,max=255"`
	DisplayLabel string  `json:"display_label" binding:"required,min=1,max=255"`
	ColorHex     *string `json:"color_hex"     binding:"omitempty,len=7"`
	SortOrder    int32   `json:"sort_order"`
	CreatedBy    *uint32 `json:"created_by"`
}

type UpdateAttributeValueRequest struct {
	AttributeNo  uint32  `json:"attribute_no"  binding:"required"`
	Value        string  `json:"value"         binding:"required,min=1,max=255"`
	DisplayLabel string  `json:"display_label" binding:"required,min=1,max=255"`
	ColorHex     *string `json:"color_hex"     binding:"omitempty,len=7"`
	SortOrder    int32   `json:"sort_order"`
	UpdatedBy    *uint32 `json:"updated_by"`
}