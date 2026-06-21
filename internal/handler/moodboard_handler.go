package handler

import (
	"net/http"
	"strconv"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type MoodboardHandler struct {
	svc        service.MoodboardService
	settingSvc service.SettingService
}

func NewMoodboardHandler(svc service.MoodboardService, settingSvc service.SettingService) *MoodboardHandler {
	return &MoodboardHandler{
		svc:        svc,
		settingSvc: settingSvc,
	}
}

// getUserEmail extracts authenticated email or fallback
func (h *MoodboardHandler) getUserEmail(c *gin.Context) string {
	if val, exists := c.Get("email"); exists {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return "CS / Designer / Estimator"
}

// GET /api/moodboards
func (h *MoodboardHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Data moodboard berhasil dimuat", data)
}

// POST /api/orders/:id/moodboard/response
func (h *MoodboardHandler) ResponseMoodboard(c *gin.Context) {
	orderID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Order ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.ResponseMoodboard(c.Request.Context(), orderID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil response moodboard order", data)
}

// POST /api/moodboards/upload-kasar
func (h *MoodboardHandler) UploadKasar(c *gin.Context) {
	mbIDStr := c.PostForm("moodboard_id")
	mbID, err := strconv.ParseUint(mbIDStr, 10, 64)
	if err != nil {
		helper.BadRequest(c, "moodboard_id tidak valid", nil)
		return
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		helper.BadRequest(c, "Tidak ada file yang diunggah", nil)
		return
	}

	files := form.File["moodboard_kasar[]"]
	if len(files) == 0 {
		files = form.File["moodboard_kasar"]
	}

	if len(files) == 0 {
		helper.BadRequest(c, "moodboard_kasar[] atau moodboard_kasar wajib diunggah", nil)
		return
	}

	data, err := h.svc.UploadKasar(c.Request.Context(), uint(mbID), files)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil mengunggah desain kasar", data)
}

// POST /api/moodboards/:id/accept-desain
func (h *MoodboardHandler) AcceptDesain(c *gin.Context) {
	mbID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	var req dto.AcceptDesainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.AcceptDesain(c.Request.Context(), mbID, req, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Opsi desain kasar & RAB berhasil disetujui", data)
}

// POST /api/moodboards/:id/revise
func (h *MoodboardHandler) Revise(c *gin.Context) {
	mbID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	var req dto.ReviseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.ReviseDesain(c.Request.Context(), mbID, req, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Status revisi desain kasar berhasil disimpan", data)
}

// DELETE /api/moodboards/files/:file_id
func (h *MoodboardHandler) DeleteFile(c *gin.Context) {
	fileID, err := helper.ParseIDParam(c, "file_id")
	if err != nil {
		helper.BadRequest(c, "File ID tidak valid", nil)
		return
	}

	if err := h.svc.DeleteFile(c.Request.Context(), fileID); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "File berhasil dihapus dari sistem", nil)
}

// ==========================================
// MODULE B: ESTIMASI (RAB KASAR)
// ==========================================

// POST /api/moodboards/:id/estimasi/response
func (h *MoodboardHandler) ResponseEstimasi(c *gin.Context) {
	mbID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Moodboard ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.ResponseEstimasi(c.Request.Context(), mbID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil response estimasi order", data)
}

// POST /api/estimasi/upload
func (h *MoodboardHandler) UploadEstimasi(c *gin.Context) {
	estIDStr := c.PostForm("estimasi_id")
	estID, err := strconv.ParseUint(estIDStr, 10, 64)
	if err != nil {
		helper.BadRequest(c, "estimasi_id tidak valid", nil)
		return
	}

	mbFileIDStr := c.PostForm("moodboard_file_id")
	mbFileID, err := strconv.ParseUint(mbFileIDStr, 10, 64)
	if err != nil {
		helper.BadRequest(c, "moodboard_file_id tidak valid", nil)
		return
	}

	file, err := c.FormFile("estimated_cost")
	if err != nil {
		helper.BadRequest(c, "estimated_cost file wajib diunggah", nil)
		return
	}

	data, err := h.svc.UploadEstimasi(c.Request.Context(), uint(estID), uint(mbFileID), file)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil mengunggah file RAB / Estimasi", data)
}

// ==========================================
// MODULE C: COMMITMENT FEE
// ==========================================

// POST /api/moodboards/:id/commitment-fee/response
func (h *MoodboardHandler) ResponseCommitmentFee(c *gin.Context) {
	mbID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Moodboard ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.ResponseCommitmentFee(c.Request.Context(), mbID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil response commitment fee order", data)
}

// PUT /api/commitment-fees/:id/total
func (h *MoodboardHandler) UpdateTotalFee(c *gin.Context) {
	feeID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	var req dto.UpdateTotalFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	data, err := h.svc.UpdateTotalFee(c.Request.Context(), feeID, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Jumlah total biaya commitment fee berhasil disimpan", data)
}

// POST /api/commitment-fees/:id/payment
func (h *MoodboardHandler) UploadPaymentProof(c *gin.Context) {
	feeID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	file, err := c.FormFile("payment_proof")
	if err != nil {
		helper.BadRequest(c, "payment_proof file wajib diunggah", nil)
		return
	}

	data, err := h.svc.UploadPaymentProof(c.Request.Context(), feeID, file)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Bukti pembayaran berhasil diunggah", data)
}

// POST /api/commitment-fees/:id/verify
func (h *MoodboardHandler) VerifyPayment(c *gin.Context) {
	feeID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.VerifyPayment(c.Request.Context(), feeID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Pembayaran commitment fee berhasil diverifikasi", data)
}

// POST /api/commitment-fees/:id/reset
func (h *MoodboardHandler) ResetPayment(c *gin.Context) {
	feeID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	data, err := h.svc.ResetPayment(c.Request.Context(), feeID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Pembayaran commitment fee berhasil direset", data)
}

// POST /api/commitment-fees/:id/revise
func (h *MoodboardHandler) RevisePaymentFee(c *gin.Context) {
	feeID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	var req dto.UpdateTotalFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	data, err := h.svc.RevisePaymentFee(c.Request.Context(), feeID, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Biaya commitment fee berhasil direvisi", data)
}

// GET /api/commitment-fees/:id/print
func (h *MoodboardHandler) PrintInvoice(c *gin.Context) {
	feeID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	pdfBytes, filename, err := h.svc.PrintInvoice(c.Request.Context(), feeID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
