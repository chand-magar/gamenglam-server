package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/services"
)

type ProductVariantHandler struct {
	svc    services.ProductVariantService
	vavSvc services.VariantAttributeValueService // ← add
}

func NewProductVariantHandler(
	svc services.ProductVariantService,
	vavSvc services.VariantAttributeValueService, // ← add
) *ProductVariantHandler {
	return &ProductVariantHandler{svc: svc, vavSvc: vavSvc}
}

func (h *ProductVariantHandler) FindAll(c *gin.Context) {
	datas, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}
	SuccessResponse(c, datas, "Records retrieved successfully.")
}

func (h *ProductVariantHandler) FindOne(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	data, err := h.svc.FindOne(c.Request.Context(), Id)
	if err != nil {
		if errors.Is(err, services.ErrProductVariantNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}
	SuccessResponse(c, data, "Record retrieved successfully.")
}

func (h *ProductVariantHandler) Create(c *gin.Context) {
	var req models.CreateProductVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	data, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create the record.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Record created successfully.",
		"data":    data,
	})
}

func (h *ProductVariantHandler) Update(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	var req models.UpdateProductVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	data, err := h.svc.Update(c.Request.Context(), Id, &req)
	if err != nil {

		log.Printf("producVariantHandler.Update error: %v", err) // ← add this
		if errors.Is(err, services.ErrProductVariantNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update the record.")
		return
	}
	SuccessResponse(c, data, "Record updated successfully.")
}

func (h *ProductVariantHandler) Delete(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), Id); err != nil {
		if errors.Is(err, services.ErrProductVariantNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete the record.")
		return
	}
	SuccessResponse(c, nil, "Record deleted successfully.")
}

func (h *ProductVariantHandler) FindByProductId(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	datas, err := h.svc.FindByProductId(c.Request.Context(), Id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}
	SuccessResponse(c, datas, "Records retrieved successfully.")
}

// POST /api/v1/product-variants/bulk
func (h *ProductVariantHandler) BulkCreate(c *gin.Context) {
	var req models.BulkCreateVariantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.svc.BulkCreate(c.Request.Context(), &req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create variants.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": fmt.Sprintf("%d variants created successfully.", result.Created),
		"data":    result,
	})
}

// GET /api/v1/product-variants/:id/attribute-values
func (h *ProductVariantHandler) GetAttributeValues(c *gin.Context) {
	variantId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	data, err := h.vavSvc.FindByVariantId(c.Request.Context(), variantId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, data, "Records retrieved successfully.")
}

// POST /api/v1/variant-images/primary-batch
func (h *VariantImageHandler) GetPrimaryBatch(c *gin.Context) {
	var req struct {
		VariantIds []string `json:"variant_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	variantIds := make([]uuid.UUID, 0, len(req.VariantIds))
	for _, id := range req.VariantIds {
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		variantIds = append(variantIds, parsed)
	}

	result, err := h.repo.GetPrimaryByVariantIds(c.Request.Context(), variantIds)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, result, "Primary images retrieved successfully.")
}

// GET /api/v1/products/:id/variants-with-attributes
func (h *ProductVariantHandler) FindByProductIdWithAttributes(c *gin.Context) {
	productId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid product ID.")
		return
	}

	data, err := h.svc.FindByProductIdWithAttributes(c.Request.Context(), productId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, data, "Records retrieved successfully.")
}

// GET /api/v1/product-variants/listing?category_id=UUID
func (h *ProductVariantHandler) GetListing(c *gin.Context) {
	var categoryId *uuid.UUID
	if idStr := c.Query("category_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid category_id.")
			return
		}
		categoryId = &id
	}

	// Accept comma-separated attribute values e.g. ?values=red,golden_and_white
	var attributeValues []string
	if valStr := c.Query("values"); valStr != "" {
		attributeValues = strings.Split(valStr, ",")
	}

	data, err := h.svc.GetListing(c.Request.Context(), categoryId, attributeValues)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, data, "Records retrieved successfully.")
}
