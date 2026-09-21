package controllers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/services"
)

type ProductHandler struct {
	svc     services.ProductService
	baseURL string
}

func NewProductHandler(svc services.ProductService, baseURL string) *ProductHandler {
	return &ProductHandler{svc: svc, baseURL: baseURL}
}

func (h *ProductHandler) FindAll(c *gin.Context) {
	categoryIdStr := c.Query("category_id")

	var datas []models.Product
	var err error

	if categoryIdStr != "" {
		categoryId, parseErr := uuid.Parse(categoryIdStr)
		if parseErr != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid category_id.")
			return
		}
		datas, err = h.svc.FindByCategory(c.Request.Context(), categoryId)
	} else {
		datas, err = h.svc.FindAll(c.Request.Context())
	}

	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, datas, "Records retrieved successfully.")
}

func (h *ProductHandler) FindOne(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	data, err := h.svc.FindOne(c.Request.Context(), Id)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, data, "Record retrieved successfully.")
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	data, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		log.Printf("productHandler.Create error: %v", err) // ← add this
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create the record.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Record created successfully.",
		"data":    data,
	})
}

func (h *ProductHandler) Update(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	data, err := h.svc.Update(c.Request.Context(), Id, &req)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update the record.")
		return
	}

	SuccessResponse(c, data, "Record updated successfully.")
}

func (h *ProductHandler) Delete(c *gin.Context) {
	Id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid ID provided.")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), Id); err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			ErrorResponse(c, http.StatusNotFound, "Record not found.")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete the record.")
		return
	}

	SuccessResponse(c, nil, "Record deleted successfully.")
}

// POST /api/v1/products/:id/image
func (h *ProductHandler) UploadImage(c *gin.Context) {
	productId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid product ID.")
		return
	}

	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "No image provided.")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		ErrorResponse(c, http.StatusBadRequest, "File type not allowed.")
		return
	}

	if fileHeader.Size > 5<<20 {
		ErrorResponse(c, http.StatusBadRequest, "File exceeds 5MB limit.")
		return
	}

	uploadDir := "./uploads/products"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create upload directory.")
		return
	}

	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to save file.")
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to write file.")
		return
	}

	imageURL := fmt.Sprintf("%s/uploads/products/%s", h.baseURL, filename)

	product, err := h.svc.UpdateImage(c.Request.Context(), productId, imageURL)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update product image.")
		return
	}

	SuccessResponse(c, product, "Image uploaded successfully.")
}

// GET /api/v1/products/:id/attributes
func (h *ProductHandler) FindAttributes(c *gin.Context) {
	productId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid product ID.")
		return
	}

	data, err := h.svc.FindAttributes(c.Request.Context(), productId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, data, "Records retrieved successfully.")
}
