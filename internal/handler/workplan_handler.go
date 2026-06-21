package handler

import (
	"net/http"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkplanHandler struct {
	svc        service.WorkplanService
	settingSvc service.SettingService
	db         *gorm.DB
	cfg        *config.Config
}

func NewWorkplanHandler(
	svc service.WorkplanService,
	settingSvc service.SettingService,
	db *gorm.DB,
	cfg *config.Config,
) *WorkplanHandler {
	return &WorkplanHandler{
		svc:        svc,
		settingSvc: settingSvc,
		db:         db,
		cfg:        cfg,
	}
}

func (h *WorkplanHandler) getUserEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return "CS / Designer / Supervisor"
	}
	return email.(string)
}

// GET /api/workplans
func (h *WorkplanHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat daftar workplan: "+err.Error())
		return
	}
	helper.OK(c, "Daftar workplan berhasil dimuat", data)
}

// GET /api/workplans/order/:orderId
func (h *WorkplanHandler) Show(c *gin.Context) {
	orderID, err := helper.ParseIDParam(c, "orderId")
	if err != nil {
		helper.BadRequest(c, "Order ID tidak valid", nil)
		return
	}

	data, err := h.svc.GetByOrderID(c.Request.Context(), orderID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Detail workplan order berhasil dimuat", data)
}

// POST /api/workplans/:id/response
// Di sini :id adalah Order ID
func (h *WorkplanHandler) Response(c *gin.Context) {
	orderID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Order ID tidak valid", nil)
		return
	}

	enabled, _ := h.settingSvc.IsEnabled(c.Request.Context(), "response_enabled")
	if !enabled {
		helper.Forbidden(c, "Fitur response dinonaktifkan")
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.Response(c.Request.Context(), orderID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil response workplan order", data)
}

// PUT /api/workplans/:id
// Di sini :id adalah Workplan ID
func (h *WorkplanHandler) Update(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	var req dto.UpdateWorkplanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.Update(c.Request.Context(), wpID, req, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Workplan berhasil diperbarui", data)
}

// GET /api/workplans/:id/excel
// Di sini :id adalah Workplan ID
func (h *WorkplanHandler) ExportExcel(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	excelBytes, filename, err := h.svc.ExportExcel(c.Request.Context(), wpID)
	if err != nil {
		helper.InternalError(c, "Gagal mengekspor Excel workplan: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}

// GET /api/settings/workplan-stages
func (h *WorkplanHandler) GetStageMasters(c *gin.Context) {
	data, err := h.svc.GetStageMasters(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat templat tahapan: "+err.Error())
		return
	}
	helper.OK(c, "Templat tahapan berhasil dimuat", data)
}

// PUT /api/settings/workplan-stages
func (h *WorkplanHandler) UpdateStageMasters(c *gin.Context) {
	var req []dto.UpdateWorkplanStageMasterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	err := h.svc.UpdateStageMasters(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Templat tahapan berhasil diperbarui", nil)
}

// POST /api/workplans/stages/:id/complete
func (h *WorkplanHandler) CompleteStage(c *gin.Context) {
	stageID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Stage ID tidak valid", nil)
		return
	}

	var req dto.CompleteStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	err = h.svc.CompleteStage(c.Request.Context(), stageID, req, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Tahapan berhasil diselesaikan", nil)
}

// POST /api/workplans/:id/request-extension
func (h *WorkplanHandler) RequestExtension(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	var req dto.RequestExtensionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	data, err := h.svc.RequestExtension(c.Request.Context(), wpID, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Pengajuan perpanjangan timeline berhasil dikirim", data)
}

// POST /api/workplans/:id/handle-extension
func (h *WorkplanHandler) HandleExtension(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	var req dto.HandleExtensionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	data, err := h.svc.HandleExtension(c.Request.Context(), wpID, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Keputusan perpanjangan timeline berhasil disimpan", data)
}

// GET /api/workplans/:id/progress/pdf
func (h *WorkplanHandler) ExportProgressPDF(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	pdfBytes, filename, err := h.svc.ExportProgressPDF(c.Request.Context(), wpID)
	if err != nil {
		helper.InternalError(c, "Gagal mengekspor PDF progress proyek: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GET /api/workplans/:id/progress/excel
func (h *WorkplanHandler) ExportProgressExcel(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	excelBytes, filename, err := h.svc.ExportProgressExcel(c.Request.Context(), wpID)
	if err != nil {
		helper.InternalError(c, "Gagal mengekspor Excel progress proyek: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}

// POST /api/workplans/stages/:stageId/defects
func (h *WorkplanHandler) ReportDefect(c *gin.Context) {
	stageID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Stage ID tidak valid", nil)
		return
	}

	var req dto.ReportDefectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.ReportDefect(c.Request.Context(), stageID, req, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Laporan defect berhasil dikirim", data)
}

// PUT /api/workplans/defects/:defectId/fix
func (h *WorkplanHandler) SubmitDefectFix(c *gin.Context) {
	defectID, err := helper.ParseIDParam(c, "defectId")
	if err != nil {
		helper.BadRequest(c, "Defect ID tidak valid", nil)
		return
	}

	var req dto.SubmitDefectFixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.SubmitDefectFix(c.Request.Context(), defectID, req, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Perbaikan defect berhasil dikirim", data)
}

// PUT /api/workplans/defects/:defectId/review
func (h *WorkplanHandler) ReviewDefect(c *gin.Context) {
	defectID, err := helper.ParseIDParam(c, "defectId")
	if err != nil {
		helper.BadRequest(c, "Defect ID tidak valid", nil)
		return
	}

	var req dto.ReviewDefectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.ReviewDefect(c.Request.Context(), defectID, req, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Review defect berhasil disimpan", data)
}

// GET /api/workplans/:id/defects
func (h *WorkplanHandler) GetDefects(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	data, err := h.svc.GetDefectsByWorkplan(c.Request.Context(), wpID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Daftar defect berhasil dimuat", data)
}

// POST /api/workplans/:id/bast
func (h *WorkplanHandler) SubmitBast(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	var req dto.SubmitBastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.SubmitBast(c.Request.Context(), wpID, req, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Data BAST berhasil disimpan", data)
}

// GET /api/workplans/:id/bast/pdf
func (h *WorkplanHandler) GenerateBastPDF(c *gin.Context) {
	wpID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Workplan ID tidak valid", nil)
		return
	}

	pdfBytes, filename, err := h.svc.GenerateBastPDF(c.Request.Context(), wpID)
	if err != nil {
		helper.InternalError(c, "Gagal men-generate PDF BAST: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.Data(200, "application/pdf", pdfBytes)
}

