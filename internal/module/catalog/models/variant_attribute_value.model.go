package models

import (
	"time"

	"github.com/google/uuid"
)

type VariantAttributeValue struct {
	VariantAttributeValueNo uint32    `json:"variant_attribute_value_no"`
	VariantAttributeValueId uuid.UUID `json:"variant_attribute_value_id"`
	VariantNo               uint32    `json:"variant_no"`
	AttributeValueNo        uint32    `json:"attribute_value_no"`
	CreatedAt               time.Time `json:"created_at"`
	CreatedBy               *uint32   `json:"created_by"`
}

type CreateVariantAttributeValueRequest struct {
	VariantNo        uint32  `json:"variant_no"         binding:"required"`
	AttributeValueNo uint32  `json:"attribute_value_no" binding:"required"`
	CreatedBy        *uint32 `json:"created_by"`
}

type UpdateVariantAttributeValueRequest struct {
	VariantNo        uint32  `json:"variant_no"         binding:"required"`
	AttributeValueNo uint32  `json:"attribute_value_no" binding:"required"`
	Status        ProductStatus `json:"status"`
	UpdatedBy     *uint32       `json:"updated_by"`
}

type DeleteVariantAttributeValueRequest struct {
	VariantNo        uint32 `json:"variant_no"         binding:"required"`
	AttributeValueNo uint32 `json:"attribute_value_no" binding:"required"`
}