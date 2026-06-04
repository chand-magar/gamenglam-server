package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/order/models"
	"github.com/chand-magar/gamenglam-server/internal/module/order/services"
)

type OrderHandler struct {
	svc services.OrderService
}

func NewOrderHandler(svc services.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req models.PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.svc.PlaceOrder(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrAddressNotFound):
			ErrorResponse(c, http.StatusNotFound, "Delivery address not found.")
		case errors.Is(err, services.ErrVariantNotFound):
			ErrorResponse(c, http.StatusNotFound, "Product variant not found.")
		case errors.Is(err, services.ErrInsufficientStock):
			ErrorResponse(c, http.StatusConflict, "Insufficient stock for one or more items.")
		default:
			ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Order placed successfully.",
		"data":    order,
	})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	customerNo, err := strconv.ParseUint(c.Param("customer_no"), 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid customer_no.")
		return
	}

	orders, err := h.svc.GetOrders(c.Request.Context(), uint32(customerNo))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, orders, "Records retrieved successfully.")
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderId, err := uuid.Parse(c.Param("order_id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid order_id.")
		return
	}

	order, err := h.svc.GetOrder(c.Request.Context(), orderId)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Order not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, order, "Record retrieved successfully.")
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	orderId, err := uuid.Parse(c.Param("order_id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid order_id.")
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.svc.UpdateStatus(c.Request.Context(), orderId, &req)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Order not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, order, "Record updated successfully.")
}

func (h *OrderHandler) GetStatusHistory(c *gin.Context) {
	orderNo, err := strconv.ParseUint(c.Param("order_no"), 10, 32)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid order_no.")
		return
	}

	history, err := h.svc.GetStatusHistory(c.Request.Context(), uint32(orderNo))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, history, "Records retrieved successfully.")
}