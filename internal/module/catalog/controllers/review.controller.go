package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/services"
)

type ReviewHandler struct {
	svc services.ReviewService
}

func NewReviewHandler(svc services.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// GET /api/v1/products/:id/reviews
func (h *ReviewHandler) FindByProduct(c *gin.Context) {
	productId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid product ID.")
		return
	}

	reviews, err := h.svc.GetByProductId(c.Request.Context(), productId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, reviews, "Records retrieved successfully.")
}

// POST /api/v1/products/:id/reviews
func (h *ReviewHandler) Create(c *gin.Context) {
	productId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid product ID.")
		return
	}

	claims, err := extractCustomerClaims(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	review, err := h.svc.Create(c.Request.Context(), productId, claims.CustomerNo, &req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to submit review.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Review submitted successfully.",
		"data":    review,
	})
}

// GET /api/v1/reviews  (admin)
func (h *ReviewHandler) FindAll(c *gin.Context) {
	page, _   := strconv.Atoi(c.DefaultQuery("page",  "1"))
	limit, _  := strconv.Atoi(c.DefaultQuery("limit", "10"))
	rating, _ := strconv.Atoi(c.DefaultQuery("rating", "0"))

	req := &models.ReviewListRequest{
		Page:   page,
		Limit:  limit,
		Search: c.Query("search"),
		Rating: rating,
		Status: c.Query("status"),
	}

	result, err := h.svc.FindAll(c.Request.Context(), req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch reviews.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// DELETE /api/v1/reviews/:id  (admin)
func (h *ReviewHandler) DeleteOne(c *gin.Context) {
	reviewId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid review ID.")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), reviewId); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete review.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review deleted."})
}
