package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc        service.OrderService
	settingSvc service.SettingService
}

func NewOrderHandler(svc service.OrderService, settingSvc service.SettingService) *OrderHandler {
	return &OrderHandler{svc: svc, settingSvc: settingSvc}
}

func (h *OrderHandler) Index(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	data, err := h.svc.GetAll(c.Request.Context(), search, status)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Data order berhasil dimuat", data)
}

func (h *OrderHandler) Store(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.Created(c, "Order berhasil dibuat", data)
}

func (h *OrderHandler) Show(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	data, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		helper.NotFound(c, "Order tidak ditemukan")
		return
	}
	helper.OK(c, "Detail order", data)
}

func (h *OrderHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Order berhasil diupdate", data)
}

func (h *OrderHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Order berhasil dihapus", nil)
}


func (h *OrderHandler) SyncTeams(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.SyncTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.SyncTeams(c.Request.Context(), id, req.UserIDs)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Tim berhasil disinkronkan", data)
}

func (h *OrderHandler) GetTeams(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	data, err := h.svc.GetTeams(c.Request.Context(), id)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Data tim order", data)
}

func (h *OrderHandler) ExportPDF(c *gin.Context) {
	params := map[string]string{
		"search":         c.Query("search"),
		"status":         c.Query("status"),
		"tahapan_proyek": c.Query("tahapan_proyek"),
		"payment_status": c.Query("payment_status"),
		"priority_level": c.Query("priority_level"),
		"jenis_interior": c.Query("jenis_interior"),
		"start_date":     c.Query("start_date"),
		"end_date":       c.Query("end_date"),
	}

	pdfBytes, filename, err := h.svc.ExportPDF(c.Request.Context(), params)
	if err != nil {
		helper.InternalError(c, "Gagal men-generate PDF Order: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.Data(200, "application/pdf", pdfBytes)
}

func (h *OrderHandler) ExportExcel(c *gin.Context) {
	params := map[string]string{
		"search":         c.Query("search"),
		"status":         c.Query("status"),
		"tahapan_proyek": c.Query("tahapan_proyek"),
		"payment_status": c.Query("payment_status"),
		"priority_level": c.Query("priority_level"),
		"jenis_interior": c.Query("jenis_interior"),
		"start_date":     c.Query("start_date"),
		"end_date":       c.Query("end_date"),
	}

	excelBytes, filename, err := h.svc.ExportExcel(c.Request.Context(), params)
	if err != nil {
		helper.InternalError(c, "Gagal men-generate Excel Order: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}
