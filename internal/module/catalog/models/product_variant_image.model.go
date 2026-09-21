package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariantImage struct {
	ImageNo   uint32    `json:"image_no"`
	ImageId   uuid.UUID `json:"image_id"`
	VariantNo uint32    `json:"variant_no"`
	ImageURL  string    `json:"image_url"`
	SortOrder int32     `json:"sort_order"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateVariantImageRequest struct {
	VariantNo uint32 `json:"variant_no" binding:"required"`
	ImageURL  string `json:"image_url"  binding:"required,max=1024"`
	SortOrder int32  `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

type UpdateVariantImageRequest struct {
	SortOrder int32 `json:"sort_order"`
	IsPrimary bool  `json:"is_primary"`
}

type BulkSaveVariantImagesRequest struct {
	VariantNo uint32   `json:"variant_no" binding:"required"`
	Images    []string `json:"images"     binding:"required,min=1"`
}
