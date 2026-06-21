package handler

import (
	"errors"
	"net/http"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InputItemHandler struct {
	svc service.InputItemService
}

func NewInputItemHandler(svc service.InputItemService) *InputItemHandler {
	return &InputItemHandler{svc: svc}
}

// GET /api/input-items
func (h *InputItemHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat rincian item: "+err.Error())
		return
	}
	helper.OK(c, "Daftar rincian item", data)
}

// GET /api/input-items/:id
func (h *InputItemHandler) Show(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	data, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound(c, "Rincian item tidak ditemukan")
			return
		}
		helper.InternalError(c, "Gagal memuat rincian item: "+err.Error())
		return
	}
	helper.OK(c, "Detail rincian item", data)
}

// GET /api/input-items/desain-final/:df_id
func (h *InputItemHandler) ShowByDesainFinal(c *gin.Context) {
	dfID, err := helper.ParseIDParam(c, "df_id")
	if err != nil {
		helper.BadRequest(c, "Desain Final ID tidak valid", nil)
		return
	}
	data, err := h.svc.GetByDesainFinalID(c.Request.Context(), dfID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Belum ada rincian item"})
			return
		}
		helper.InternalError(c, "Gagal memuat rincian item: "+err.Error())
		return
	}
	helper.OK(c, "Detail rincian item berdasarkan desain final", data)
}

// POST /api/input-items
func (h *InputItemHandler) Store(c *gin.Context) {
	var req dto.CreateInputItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, "Gagal menyimpan rincian item: "+err.Error())
		return
	}
	helper.Created(c, "Rincian item berhasil dibuat", data)
}

// PUT /api/input-items/:id
func (h *InputItemHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.UpdateInputItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.InternalError(c, "Gagal memperbarui rincian item: "+err.Error())
		return
	}
	helper.OK(c, "Rincian item berhasil diperbarui", data)
}

// DELETE /api/input-items/:id
func (h *InputItemHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.InternalError(c, "Gagal menghapus rincian item: "+err.Error())
		return
	}
	helper.OK(c, "Rincian item berhasil dihapus", nil)
}

func (h *InputItemHandler) getUserEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return "CS / Designer / Estimator"
	}
	return email.(string)
}

// POST /api/input-items/desain-final/:df_id/response/designer
func (h *InputItemHandler) InputItemResponseDesigner(c *gin.Context) {
	dfID, err := helper.ParseIDParam(c, "df_id")
	if err != nil {
		helper.BadRequest(c, "Desain Final ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.InputItemResponseDesigner(c.Request.Context(), dfID, userEmail)
	if err != nil {
		helper.InternalError(c, "Gagal memberikan response desainer: "+err.Error())
		return
	}
	helper.OK(c, "Tanggapan desainer berhasil disimpan", data)
}

// POST /api/input-items/desain-final/:df_id/response/marketing
func (h *InputItemHandler) InputItemResponseMarketing(c *gin.Context) {
	dfID, err := helper.ParseIDParam(c, "df_id")
	if err != nil {
		helper.BadRequest(c, "Desain Final ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.InputItemResponseMarketing(c.Request.Context(), dfID, userEmail)
	if err != nil {
		helper.InternalError(c, "Gagal memberikan response marketing: "+err.Error())
		return
	}
	helper.OK(c, "Tanggapan marketing berhasil disimpan", data)
}
