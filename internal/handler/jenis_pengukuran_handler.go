package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type JenisPengukuranHandler struct {
	svc service.JenisPengukuranService
}

func NewJenisPengukuranHandler(svc service.JenisPengukuranService) *JenisPengukuranHandler {
	return &JenisPengukuranHandler{svc: svc}
}

func (h *JenisPengukuranHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat jenis pengukuran")
		return
	}
	helper.OK(c, "Daftar jenis pengukuran", data)
}

func (h *JenisPengukuranHandler) Store(c *gin.Context) {
	var req dto.CreateJenisPengukuranRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.Created(c, "Jenis pengukuran berhasil dibuat", data)
}

func (h *JenisPengukuranHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.UpdateJenisPengukuranRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Jenis pengukuran berhasil diupdate", data)
}

func (h *JenisPengukuranHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Jenis pengukuran berhasil dihapus", nil)
}
