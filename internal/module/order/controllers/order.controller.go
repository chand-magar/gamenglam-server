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

// GET /api/v1/orders
func (h *OrderHandler) FindAll(c *gin.Context) {
	page, _  := strconv.Atoi(c.DefaultQuery("page",  "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	req := &models.OrderListRequest{
		Page:   page,
		Limit:  limit,
		Status: c.Query("order_status"),
		Search: c.Query("search"),
	}

	result, err := h.svc.FindAll(c.Request.Context(), req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Records retrieved successfully.",
		"data":        result.Orders,
		"total":       result.Total,
		"page":        result.Page,
		"limit":       result.Limit,
		"total_pages": result.TotalPages,
	})
}

// POST /api/v1/orders
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

// GET /api/v1/orders/customer/:id
func (h *OrderHandler) GetOrders(c *gin.Context) {
	customerId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID.")
		return
	}

	orders, err := h.svc.GetOrders(c.Request.Context(), customerId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, orders, "Records retrieved successfully.")
}

// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid order ID.")
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

// PUT /api/v1/orders/:id/status
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	orderId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid order ID.")
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

// GET /api/v1/orders/:id/history
func (h *OrderHandler) GetStatusHistory(c *gin.Context) {
	orderId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid order ID.")
		return
	}

	order, err := h.svc.GetOrder(c.Request.Context(), orderId)
	if err != nil || order == nil {
		ErrorResponse(c, http.StatusNotFound, "Order not found.")
		return
	}

	history, err := h.svc.GetStatusHistory(c.Request.Context(), order.OrderNo)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, history, "Records retrieved successfully.")
}
