package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RABHandler struct {
	svc service.RABService
}

func NewRABHandler(svc service.RABService) *RABHandler {
	return &RABHandler{svc: svc}
}

// getUserEmail extracts authenticated email or fallback
func (h *RABHandler) getUserEmail(c *gin.Context) string {
	if val, exists := c.Get("email"); exists {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return "CS / Estimator / Finance"
}

// GET /api/rab
func (h *RABHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat daftar RAB: "+err.Error())
		return
	}
	helper.OK(c, "Daftar RAB", data)
}

// GET /api/rab/:id
func (h *RABHandler) Show(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	mode := c.Query("mode")
	if mode == "" {
		mode = "internal"
	}

	// We'll modify GetByID to return the mapped calculation dynamically based on visual mode
	// But to keep it generic, we'll implement that in rab_service.go
	data, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound(c, "RAB tidak ditemukan")
			return
		}
		helper.InternalError(c, "Gagal memuat detail RAB: "+err.Error())
		return
	}
	helper.OK(c, "Detail RAB", data)
}

// GET /api/rab/input-item/:input_item_id
func (h *RABHandler) ShowByInputItemID(c *gin.Context) {
	inputItemID, err := helper.ParseIDParam(c, "input_item_id")
	if err != nil {
		helper.BadRequest(c, "Input Item ID tidak valid", nil)
		return
	}

	data, err := h.svc.GetByInputItemID(c.Request.Context(), inputItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Belum ada RAB"})
			return
		}
		helper.InternalError(c, "Gagal memuat RAB berdasarkan input item: "+err.Error())
		return
	}
	helper.OK(c, "Detail RAB berdasarkan input item", data)
}

// POST /api/rab
func (h *RABHandler) Store(c *gin.Context) {
	var req dto.CreateRABRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, "Gagal menyimpan RAB: "+err.Error())
		return
	}
	helper.Created(c, "RAB berhasil dibuat", data)
}

// PUT /api/rab/:id
func (h *RABHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	var req dto.UpdateRABRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.InternalError(c, "Gagal memperbarui RAB: "+err.Error())
		return
	}
	helper.OK(c, "RAB berhasil diperbarui", data)
}

// POST /api/rab/:id/submit
func (h *RABHandler) Submit(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.Submit(c.Request.Context(), id, userEmail)
	if err != nil {
		helper.InternalError(c, "Gagal submit RAB: "+err.Error())
		return
	}
	helper.OK(c, "RAB berhasil disubmit dan dikunci", data)
}

// DELETE /api/rab/:id
func (h *RABHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.InternalError(c, "Gagal menghapus RAB: "+err.Error())
		return
	}
	helper.OK(c, "RAB berhasil dihapus", nil)
}

// GET /api/rab/:id/export
func (h *RABHandler) Export(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	exportType := c.Query("type") // "pdf" or "excel"
	mode := c.Query("mode")       // "internal", "kontrak", "vendor", "jasa"
	if mode == "" {
		mode = "internal"
	}

	if exportType == "pdf" {
		pdfBytes, filename, err := h.svc.ExportPDF(c.Request.Context(), id, mode)
		if err != nil {
			helper.InternalError(c, "Gagal ekspor PDF: "+err.Error())
			return
		}
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
		return
	} else if exportType == "excel" {
		excelBytes, filename, err := h.svc.ExportExcel(c.Request.Context(), id, mode)
		if err != nil {
			helper.InternalError(c, "Gagal ekspor Excel: "+err.Error())
			return
		}
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Length", strconv.Itoa(len(excelBytes)))
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
		return
	}

	helper.BadRequest(c, "Parameter type ekspor tidak valid (harus pdf atau excel)", nil)
}
