package handler

import (
	"strconv"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Index(c *gin.Context) {
	search := c.Query("search")
	var roleID uint
	if rid := c.Query("role_id"); rid != "" {
		if v, err := strconv.ParseUint(rid, 10, 64); err == nil {
			roleID = uint(v)
		}
	}
	data, err := h.svc.GetAll(c.Request.Context(), search, roleID)
	if err != nil {
		helper.InternalError(c, "Gagal memuat users")
		return
	}
	helper.OK(c, "Daftar user", data)
}

func (h *UserHandler) Store(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.BadRequest(c, err.Error(), nil)
		return
	}
	helper.Created(c, "User berhasil dibuat", data)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.BadRequest(c, err.Error(), nil)
		return
	}
	helper.OK(c, "User berhasil diupdate", data)
}

func (h *UserHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	} 
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "User berhasil dihapus", nil)
}
