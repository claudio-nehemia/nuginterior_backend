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

type InvoiceHandler struct {
	svc        service.InvoiceService
	settingSvc service.SettingService
}

func NewInvoiceHandler(svc service.InvoiceService, settingSvc service.SettingService) *InvoiceHandler {
	return &InvoiceHandler{
		svc:        svc,
		settingSvc: settingSvc,
	}
}

func (h *InvoiceHandler) getUserEmail(c *gin.Context) string {
	if val, exists := c.Get("email"); exists {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return "CS / Estimator / Finance"
}

// GET /api/invoices
func (h *InvoiceHandler) Index(c *gin.Context) {
	data, err := h.svc.GetContractInvoiceList(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat daftar tagihan kontrak: "+err.Error())
		return
	}
	helper.OK(c, "Daftar tagihan kontrak", data)
}

// GET /api/invoices/contract/:contractId
func (h *InvoiceHandler) Show(c *gin.Context) {
	contractID, err := helper.ParseIDParam(c, "contractId")
	if err != nil {
		helper.BadRequest(c, "Contract ID tidak valid", nil)
		return
	}

	data, err := h.svc.GetInvoicesByContractID(c.Request.Context(), contractID)
	if err != nil {
		helper.InternalError(c, "Gagal memuat detail invoice kontrak: "+err.Error())
		return
	}
	helper.OK(c, "Detail invoice kontrak", data)
}

// POST /api/invoices/contract/:contractId/response
func (h *InvoiceHandler) Response(c *gin.Context) {
	contractID, err := helper.ParseIDParam(c, "contractId")
	if err != nil {
		helper.BadRequest(c, "Contract ID tidak valid", nil)
		return
	}

	enabled, _ := h.settingSvc.IsEnabled(c.Request.Context(), "response_enabled")
	if !enabled {
		helper.Forbidden(c, "Fitur response regular tidak diaktifkan")
		return
	}

	email := h.getUserEmail(c)
	data, err := h.svc.SubmitInvoiceResponse(c.Request.Context(), contractID, email)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil merespons invoice kontrak", data)
}

// PUT /api/invoices/:id/deadline
func (h *InvoiceHandler) UpdateDeadline(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	enabled, _ := h.settingSvc.IsEnabled(c.Request.Context(), "invoice_deadline_enabled")
	if !enabled {
		helper.Forbidden(c, "Fitur deadline invoice tidak aktif")
		return
	}

	var req dto.UpdateInvoiceDeadlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal: "+err.Error(), nil)
		return
	}

	data, err := h.svc.UpdateDeadline(c.Request.Context(), id, req.Deadline)
	if err != nil {
		helper.InternalError(c, "Gagal memperbarui deadline: "+err.Error())
		return
	}
	helper.OK(c, "Deadline tagihan berhasil diperbarui", data)
}

// POST /api/invoices/:id/payment
func (h *InvoiceHandler) UploadPayment(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	file, err := c.FormFile("payment_proof")
	if err != nil {
		helper.BadRequest(c, "Bukti pembayaran wajib diunggah", nil)
		return
	}

	data, err := h.svc.UploadPaymentProof(c.Request.Context(), id, file)
	if err != nil {
		helper.InternalError(c, "Gagal mengunggah bukti pembayaran: "+err.Error())
		return
	}
	helper.OK(c, "Bukti pembayaran berhasil diunggah, invoice Terbayar", data)
}

// GET /api/invoices/:id/pdf
func (h *InvoiceHandler) ExportPDF(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	pdfBytes, filename, err := h.svc.GenerateInvoicePDF(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound(c, "Invoice tidak ditemukan")
			return
		}
		helper.InternalError(c, "Gagal mengekspor PDF invoice: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "inline; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
