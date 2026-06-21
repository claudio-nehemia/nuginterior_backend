package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	data, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat statistik dashboard")
		return
	}
	helper.OK(c, "Statistik dashboard", data)
}
