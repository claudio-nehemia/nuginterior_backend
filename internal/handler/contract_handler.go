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

type ContractHandler struct {
	svc        service.ContractService
	settingSvc service.SettingService
}

func NewContractHandler(svc service.ContractService, settingSvc service.SettingService) *ContractHandler {
	return &ContractHandler{
		svc:        svc,
		settingSvc: settingSvc,
	}
}

func (h *ContractHandler) getUserEmail(c *gin.Context) string {
	if val, exists := c.Get("email"); exists {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return "CS / Estimator / Finance"
}

// GET /api/contracts
func (h *ContractHandler) Index(c *gin.Context) {
	data, err := h.svc.GetContractList(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat daftar kontrak: "+err.Error())
		return
	}
	helper.OK(c, "Daftar kontrak", data)
}

// POST /api/contracts/rab/:rabId/response
func (h *ContractHandler) Response(c *gin.Context) {
	rabID, err := helper.ParseIDParam(c, "rabId")
	if err != nil {
		helper.BadRequest(c, "RAB ID tidak valid", nil)
		return
	}

	enabled, _ := h.settingSvc.IsEnabled(c.Request.Context(), "response_enabled")
	if !enabled {
		helper.Forbidden(c, "Fitur response regular tidak diaktifkan")
		return
	}

	email := h.getUserEmail(c)
	data, err := h.svc.SubmitResponse(c.Request.Context(), rabID, email)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil merespons kontrak", data)
}

// POST /api/contracts
func (h *ContractHandler) Store(c *gin.Context) {
	var req dto.CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal: "+err.Error(), nil)
		return
	}

	data, err := h.svc.CreateContract(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, "Gagal membuat kontrak: "+err.Error())
		return
	}
	helper.OK(c, "Kontrak berhasil dibuat (draft)", data)
}

// GET /api/contracts/:id/pdf
func (h *ContractHandler) ExportPDF(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	pdfBytes, filename, err := h.svc.GenerateContractPDF(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound(c, "Kontrak tidak ditemukan")
			return
		}
		helper.InternalError(c, "Gagal mengekspor PDF kontrak: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "inline; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// POST /api/contracts/:id/upload-signed
func (h *ContractHandler) UploadSigned(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	file, err := c.FormFile("signed_contract")
	if err != nil {
		helper.BadRequest(c, "File signed_contract wajib diunggah", nil)
		return
	}

	data, err := h.svc.UploadSignedContract(c.Request.Context(), id, file)
	if err != nil {
		helper.InternalError(c, "Gagal mengunggah file tanda tangan kontrak: "+err.Error())
		return
	}
	helper.OK(c, "File tanda tangan kontrak berhasil diunggah, kontrak DEAL", data)
}
