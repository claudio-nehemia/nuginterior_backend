package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	svc service.PermissionService
}

func NewPermissionHandler(svc service.PermissionService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

func (h *PermissionHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAllGrouped(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat permissions")
		return
	}
	helper.OK(c, "Daftar permission", data)
}
