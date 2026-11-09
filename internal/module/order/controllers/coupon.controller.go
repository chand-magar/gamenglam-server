package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
	"github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

type CouponHandler struct {
	svc services.CouponService
}

func NewCouponHandler(svc services.CouponService) *CouponHandler {
	return &CouponHandler{svc: svc}
}

// GET /api/v1/coupons
func (h *CouponHandler) FindAll(c *gin.Context) {
	page, _  := strconv.Atoi(c.DefaultQuery("page",  "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	req := &models.CouponListRequest{
		Page:   page,
		Limit:  limit,
		Status: c.Query("status"),
		Search: c.Query("search"),
	}

	result, err := h.svc.FindAll(c.Request.Context(), req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch coupons.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GET /api/v1/coupons/:id
func (h *CouponHandler) FindOne(c *gin.Context) {
	couponId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid coupon ID.")
		return
	}

	coupon, err := h.svc.FindById(c.Request.Context(), couponId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch coupon.")
		return
	}
	if coupon == nil {
		ErrorResponse(c, http.StatusNotFound, "Coupon not found.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": coupon})
}

// POST /api/v1/coupons
func (h *CouponHandler) Create(c *gin.Context) {
	var req models.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	coupon, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create coupon.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Coupon created.", "data": coupon})
}

// PUT /api/v1/coupons/:id
func (h *CouponHandler) Update(c *gin.Context) {
	couponId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid coupon ID.")
		return
	}

	var req models.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	coupon, err := h.svc.Update(c.Request.Context(), couponId, &req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update coupon.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Coupon updated.", "data": coupon})
}

// DELETE /api/v1/coupons/:id
func (h *CouponHandler) Delete(c *gin.Context) {
	couponId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid coupon ID.")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), couponId); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete coupon.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Coupon deleted."})
}
