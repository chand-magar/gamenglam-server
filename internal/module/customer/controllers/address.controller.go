package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/customer/models"
	"github.com/chand-magar/gamenglam-server/internal/module/customer/services"
)

type AddressHandler struct {
	svc         services.AddressService
	customerSvc services.CustomerService
}

func NewAddressHandler(svc services.AddressService, customerSvc services.CustomerService) *AddressHandler {
	return &AddressHandler{svc: svc, customerSvc: customerSvc}
}

// GET /api/v1/addresses
func (h *AddressHandler) FindAll(c *gin.Context) {
	customerId, err := extractCustomerIdFromToken(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	customer, err := h.customerSvc.FindOne(c.Request.Context(), customerId)
	if err != nil || customer == nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	addresses, err := h.svc.FindAllByCustomer(c.Request.Context(), customer.CustomerNo)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, addresses, "Records retrieved successfully.")
}

// POST /api/v1/addresses
func (h *AddressHandler) Create(c *gin.Context) {
	customerId, err := extractCustomerIdFromToken(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	customer, err := h.customerSvc.FindOne(c.Request.Context(), customerId)
	if err != nil || customer == nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	var req models.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Set customer_no from token — never from request body
	req.CustomerNo = customer.CustomerNo

	address, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create address.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Address created successfully.",
		"data":    address,
	})
}

// PUT /api/v1/addresses/:id
func (h *AddressHandler) Update(c *gin.Context) {
	customerId, err := extractCustomerIdFromToken(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid address ID.")
		return
	}

	// Verify ownership
	if err := h.verifyOwnership(c, customerId, Id); err != nil {
		ErrorResponse(c, http.StatusForbidden, "You do not have permission to update this address.")
		return
	}

	var req models.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	address, err := h.svc.Update(c.Request.Context(), Id, &req)
	if err != nil {
		if errors.Is(err, services.ErrAddressNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Address not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update address.")
		return
	}

	SuccessResponse(c, address, "Address updated successfully.")
}

// DELETE /api/v1/addresses/:id
func (h *AddressHandler) Delete(c *gin.Context) {
	customerId, err := extractCustomerIdFromToken(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "Unauthorized.")
		return
	}

	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid address ID.")
		return
	}

	// Verify ownership
	if err := h.verifyOwnership(c, customerId, Id); err != nil {
		ErrorResponse(c, http.StatusForbidden, "You do not have permission to delete this address.")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), Id); err != nil {
		if errors.Is(err, services.ErrAddressNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Address not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete address.")
		return
	}

	SuccessResponse(c, nil, "Address deleted successfully.")
}

// ── verifyOwnership ───────────────────────────────────────────────────────────

func (h *AddressHandler) verifyOwnership(c *gin.Context, customerId uuid.UUID, addressId uuid.UUID) error {
	// Get customer to resolve customer_no
	customer, err := h.customerSvc.FindOne(c.Request.Context(), customerId)
	if err != nil || customer == nil {
		return errors.New("customer not found")
	}

	// Get address and check it belongs to this customer
	address, err := h.svc.FindOneById(c.Request.Context(), addressId)
	if err != nil || address == nil {
		return errors.New("address not found")
	}

	if address.CustomerNo != customer.CustomerNo {
		return errors.New("address does not belong to customer")
	}

	return nil
}
