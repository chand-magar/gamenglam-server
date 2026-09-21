package models

import (
	"time"

	"github.com/google/uuid"
)

type AttributeValue struct {
	AttributeValueNo uint32    `json:"-"`
	AttributeValueId uuid.UUID `json:"attribute_value_id"`
	AttributeNo      uint32    `json:"-"`
	AttributeId      uuid.UUID `json:"attribute_id"` // ← expose
	Value            string    `json:"value"`
	DisplayLabel     string    `json:"display_label"`
	ColorHex         *string   `json:"color_hex"`
	SortOrder        int32     `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        *uint32   `json:"-"`
	UpdatedAt        time.Time `json:"updated_at"`
	UpdatedBy        *uint32   `json:"-"`
}

type CreateAttributeValueRequest struct {
	AttributeId  uuid.UUID `json:"-"` // ← set from URL
	Value        string    `json:"value"         binding:"required,min=1,max=255"`
	DisplayLabel string    `json:"display_label" binding:"required,min=1,max=255"`
	ColorHex     *string   `json:"color_hex"     binding:"omitempty,len=7"`
	SortOrder    int32     `json:"sort_order"`
	CreatedBy    *uint32   `json:"-"`
}

type UpdateAttributeValueRequest struct {
	Value        string  `json:"value"         binding:"required,min=1,max=255"`
	DisplayLabel string  `json:"display_label" binding:"required,min=1,max=255"`
	ColorHex     *string `json:"color_hex"     binding:"omitempty,len=7"`
	SortOrder    int32   `json:"sort_order"`
	UpdatedBy    *uint32 `json:"-"`
}
