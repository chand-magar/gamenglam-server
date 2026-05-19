package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	CategoryNo   uint32     `json:"category_no"`
	CategoryId   uuid.UUID  `json:"category_id"`
	ParentNo     *uint32    `json:"parent_no"`
	CategoryName string     `json:"category_name"`
	Slug         string     `json:"slug"`
	Details      string     `json:"details"`
	Status       StatusEnum `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    *uint32    `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    *uint32    `json:"updated_by"`
}

type CreateCategoryRequest struct {
	ParentNo     *uint32 `json:"parent_no"`
	CategoryName string  `json:"category_name" binding:"required,min=2,max=255"`
	Slug         string  `json:"slug"          binding:"required,min=2,max=255"`
	Details      string  `json:"details"`
	CreatedBy    *uint32 `json:"created_by"`
}

type UpdateCategoryRequest struct {
	ParentNo     *uint32 `json:"parent_no"`
	CategoryName string  `json:"category_name" binding:"required,min=2,max=255"`
	Slug         string  `json:"slug"          binding:"required,min=2,max=255"`
	Details      string  `json:"details"`
	UpdatedBy    *uint32 `json:"updated_by"`
}