package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type TerminHandler struct {
	svc service.TerminService
}

func NewTerminHandler(svc service.TerminService) *TerminHandler {
	return &TerminHandler{svc: svc}
}

func (h *TerminHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat termin")
		return
	}
	helper.OK(c, "Daftar termin", data)
}

func (h *TerminHandler) Store(c *gin.Context) {
	var req dto.CreateTerminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.Created(c, "Termin berhasil dibuat", data)
}

func (h *TerminHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.UpdateTerminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Termin berhasil diupdate", data)
}

func (h *TerminHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Termin berhasil dihapus", nil)
}
