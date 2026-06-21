package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	svc service.RoleService
}

func NewRoleHandler(svc service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

func (h *RoleHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat roles")
		return
	}
	helper.OK(c, "Daftar role", data)
}

func (h *RoleHandler) Show(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	data, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		helper.NotFound(c, "Role tidak ditemukan")
		return
	}
	helper.OK(c, "Detail role", data)
}

func (h *RoleHandler) Store(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.Created(c, "Role berhasil dibuat", data)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Role berhasil diupdate", data)
}

func (h *RoleHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.Conflict(c, err.Error())
		return
	}
	helper.OK(c, "Role berhasil dihapus", nil)
}

func (h *RoleHandler) SyncPermissions(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.SyncPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	if err := h.svc.SyncPermissions(c.Request.Context(), id, req); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Permissions berhasil di-sync", nil)
}
