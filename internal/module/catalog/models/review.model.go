package models

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ReviewNo     uint32    `json:"-"`
	ReviewId     uuid.UUID `json:"review_id"`
	ProductNo    uint32    `json:"-"`
	ProductId    uuid.UUID `json:"product_id"`
	CustomerNo   uint32    `json:"-"`
	CustomerName string    `json:"customer_name"`
	Rating       int       `json:"rating"`
	Title        string    `json:"title,omitempty"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateReviewRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Title  string `json:"title"`
	Body   string `json:"body" binding:"required,min=10"`
}

type AdminReview struct {
	ReviewNo    uint32    `json:"-"`
	ReviewId    uuid.UUID `json:"review_id"`
	ProductNo   uint32    `json:"-"`
	ProductId   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	CustomerNo  uint32    `json:"-"`
	CustomerName string   `json:"customer_name"`
	Rating      int       `json:"rating"`
	Title       string    `json:"title,omitempty"`
	Body        string    `json:"body"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ReviewListRequest struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Search string `form:"search"`
	Rating int    `form:"rating"`
	Status string `form:"status"`
}

type ReviewListResponse struct {
	Items      []AdminReview `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}
