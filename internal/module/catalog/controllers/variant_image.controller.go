package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/chand-magar/gamenglam-server/internal/module/catalog/models"
	"github.com/chand-magar/gamenglam-server/internal/module/catalog/repository"
)

type VariantImageHandler struct {
	repo    repository.VariantImageRepository
	baseURL string
}

func NewVariantImageHandler(repo repository.VariantImageRepository, baseURL string) *VariantImageHandler {
	return &VariantImageHandler{repo: repo, baseURL: baseURL}
}

// GET /api/v1/product-variants/:id/images
func (h *VariantImageHandler) FindAll(c *gin.Context) {
	variantId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid variant ID.")
		return
	}

	images, err := h.repo.GetByVariantId(c.Request.Context(), variantId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	SuccessResponse(c, images, "Records retrieved successfully.")
}

// POST /api/v1/product-variants/:id/images/upload
func (h *VariantImageHandler) Upload(c *gin.Context) {
	variantId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid variant ID.")
		return
	}

	variantNo, err := h.repo.GetVariantNoById(c.Request.Context(), variantId)
	if err != nil || variantNo == 0 {
		ErrorResponse(c, http.StatusNotFound, "Variant not found.")
		return
	}

	// Gin handles multipart automatically — use FormFile/MultipartForm directly
	form, err := c.MultipartForm()
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Failed to read multipart form.")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		ErrorResponse(c, http.StatusBadRequest, "No images provided.")
		return
	}

	uploadDir := "./uploads/variants"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create upload directory.")
		return
	}

	var savedImages []models.ProductVariantImage

	for i, fileHeader := range files {
		// Validate extension
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		allowed := map[string]bool{
			".jpg": true, ".jpeg": true,
			".png": true, ".webp": true,
		}
		if !allowed[ext] {
			ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("File type %s not allowed.", ext))
			return
		}

		// Validate size — 5MB
		if fileHeader.Size > 5<<20 {
			ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("File %s exceeds 5MB limit.", fileHeader.Filename))
			return
		}

		// Unique filename
		filename := fmt.Sprintf("%d-%s%s",
			time.Now().UnixNano(),
			uuid.New().String()[:8],
			ext,
		)
		filePath := filepath.Join(uploadDir, filename)

		src, err := fileHeader.Open()
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to open file.")
			return
		}
		defer src.Close()

		dst, err := os.Create(filePath)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to save file.")
			return
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to write file.")
			return
		}

		imageURL := fmt.Sprintf("%s/uploads/variants/%s", h.baseURL, filename)

		img, err := h.repo.Save(c.Request.Context(), &models.ProductVariantImage{
			VariantNo: variantNo,
			ImageURL:  imageURL,
			SortOrder: int32(i),
			IsPrimary: i == 0,
		})
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to save image record.")
			return
		}

		savedImages = append(savedImages, *img)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": fmt.Sprintf("%d image(s) uploaded successfully.", len(savedImages)),
		"data":    savedImages,
	})
}

// POST /api/v1/product-variants/:id/images/url
func (h *VariantImageHandler) SaveURL(c *gin.Context) {
	variantId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid variant ID.")
		return
	}

	variantNo, err := h.repo.GetVariantNoById(c.Request.Context(), variantId)
	if err != nil || variantNo == 0 {
		ErrorResponse(c, http.StatusNotFound, "Variant not found.")
		return
	}

	var req struct {
		ImageURL  string `json:"image_url"  binding:"required"`
		IsPrimary bool   `json:"is_primary"`
		SortOrder int32  `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	img, err := h.repo.Save(c.Request.Context(), &models.ProductVariantImage{
		VariantNo: variantNo,
		ImageURL:  req.ImageURL,
		SortOrder: req.SortOrder,
		IsPrimary: req.IsPrimary,
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to save image.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Image saved successfully.",
		"data":    img,
	})
}

// DELETE /api/v1/variant-images/:id
func (h *VariantImageHandler) Delete(c *gin.Context) {
	imageId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid image ID.")
		return
	}

	img, err := h.repo.GetById(c.Request.Context(), imageId)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "An internal server error occurred.")
		return
	}

	if err := h.repo.Remove(c.Request.Context(), imageId); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete image.")
		return
	}

	// Remove file from disk if local upload
	if img != nil && strings.Contains(img.ImageURL, "/uploads/variants/") {
		filename := filepath.Base(img.ImageURL)
		localPath := filepath.Join("./uploads/variants", filename)
		_ = os.Remove(localPath)
	}

	SuccessResponse(c, nil, "Image deleted successfully.")
}
