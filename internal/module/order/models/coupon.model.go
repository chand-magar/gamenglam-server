package models

import (
	"time"

	"github.com/google/uuid"
)

type Coupon struct {
	CouponNo        uint32     `json:"-"`
	CouponId        uuid.UUID  `json:"coupon_id"`
	Code            string     `json:"code"`
	Description     string     `json:"description"`
	DiscountType    string     `json:"discount_type"`
	DiscountValue   float64    `json:"discount_value"`
	MinOrderAmount  float64    `json:"min_order_amount"`
	MaxUses         *int       `json:"max_uses"`
	UsedCount       int        `json:"used_count"`
	ValidFrom       time.Time  `json:"valid_from"`
	ValidUntil      *time.Time `json:"valid_until"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateCouponRequest struct {
	Code           string     `json:"code"            binding:"required,min=3,max=50"`
	Description    string     `json:"description"`
	DiscountType   string     `json:"discount_type"   binding:"required,oneof=percentage fixed"`
	DiscountValue  float64    `json:"discount_value"  binding:"required,gt=0"`
	MinOrderAmount float64    `json:"min_order_amount"`
	MaxUses        *int       `json:"max_uses"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until"`
}

type UpdateCouponRequest struct {
	Description    *string    `json:"description"`
	DiscountType   *string    `json:"discount_type"   binding:"omitempty,oneof=percentage fixed"`
	DiscountValue  *float64   `json:"discount_value"  binding:"omitempty,gt=0"`
	MinOrderAmount *float64   `json:"min_order_amount"`
	MaxUses        *int       `json:"max_uses"`
	ValidFrom      *time.Time `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until"`
	Status         *string    `json:"status"          binding:"omitempty,oneof=A I"`
}

type CouponListRequest struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Status string `form:"status"`
	Search string `form:"search"`
}

type CouponListResponse struct {
	Items      []Coupon `json:"items"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}
