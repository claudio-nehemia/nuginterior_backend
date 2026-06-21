package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProjectLogTaskHandler struct {
	svc service.ProjectLogTaskService
}

func NewProjectLogTaskHandler(svc service.ProjectLogTaskService) *ProjectLogTaskHandler {
	return &ProjectLogTaskHandler{svc: svc}
}

func (h *ProjectLogTaskHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat log task")
		return
	}
	helper.OK(c, "Daftar log task", data)
}
