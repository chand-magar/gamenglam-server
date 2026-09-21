package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	CategoryNo   uint32     `json:"-"`
	CategoryId   uuid.UUID  `json:"category_id"`
	ParentNo     *uint32    `json:"-"`         // ← expose for filtering
	ParentId     *uuid.UUID `json:"parent_id"` // ← expose for frontend
	CategoryName string     `json:"category_name"`
	Slug         string     `json:"slug"`
	Details      string     `json:"details"`
	Status       StatusEnum `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    *uint32    `json:"-"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    *uint32    `json:"-"`
}
type CreateCategoryRequest struct {
	ParentId     *uuid.UUID `json:"parent_id"` // ← pointer so it can be null
	CategoryName string     `json:"category_name" binding:"required,min=2,max=255"`
	Slug         string     `json:"slug"          binding:"required,min=2,max=255"`
	Details      string     `json:"details"`
	CreatedBy    *uint32    `json:"-"`
}

type UpdateCategoryRequest struct {
	ParentId     *uuid.UUID `json:"parent_id"` // ← pointer so it can be null
	CategoryName string     `json:"category_name" binding:"required,min=2,max=255"`
	Slug         string     `json:"slug"          binding:"required,min=2,max=255"`
	Details      string     `json:"details"`
	UpdatedBy    *uint32    `json:"-"`
}
