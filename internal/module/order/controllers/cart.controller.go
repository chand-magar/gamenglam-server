package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
	"github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

type CartHandler struct {
	svc services.CartService
}

func NewCartHandler(svc services.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	customerId, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid customer_id.")
		return
	}

	cart, err := h.svc.GetCart(c.Request.Context(), customerId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, cart, "Record retrieved successfully.")
}

func (h *CartHandler) AddItem(c *gin.Context) {
	customerId, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid customer_id.")
		return
	}

	var req models.AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.AddItem(c.Request.Context(), customerId, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrVariantNotFound):
			ErrorResponse(c, http.StatusNotFound, "Product variant not found.")
		case errors.Is(err, services.ErrInsufficientStock):
			ErrorResponse(c, http.StatusConflict, "Insufficient stock.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Item added to cart successfully.",
		"data":    item,
	})
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	cartItemId, err := uuid.Parse(c.Param("cart_item_id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid cart_item_id.")
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.svc.UpdateItem(c.Request.Context(), cartItemId, &req)
	if err != nil {
		if errors.Is(err, services.ErrCartItemNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Cart item not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, item, "Record updated successfully.")
}

func (h *CartHandler) RemoveItem(c *gin.Context) {
	cartItemId, err := uuid.Parse(c.Param("cart_item_id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid cart_item_id.")
		return
	}

	if err := h.svc.RemoveItem(c.Request.Context(), cartItemId); err != nil {
		if errors.Is(err, services.ErrCartItemNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Cart item not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, nil, "Record deleted successfully.")
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	customerId, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid customer_id.")
		return
	}

	if err := h.svc.ClearCart(c.Request.Context(), customerId); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, nil, "Cart cleared successfully.")
}