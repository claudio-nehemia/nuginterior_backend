package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	cfg *config.Config
}

func NewUploadHandler(cfg *config.Config) *UploadHandler {
	return &UploadHandler{cfg: cfg}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		helper.BadRequest(c, "Image is required", nil)
		return
	}

	// Create upload dir if not exists
	if _, err := os.Stat(h.cfg.UploadDir); os.IsNotExist(err) {
		os.MkdirAll(h.cfg.UploadDir, os.ModePerm)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(h.cfg.UploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		helper.InternalError(c, "Failed to save file")
		return
	}

	// Return the relative URL (assuming the server serves the uploads directory)
	fileURL := fmt.Sprintf("/uploads/%s", filename)
	helper.OK(c, "Image uploaded successfully", gin.H{"url": fileURL})
}
