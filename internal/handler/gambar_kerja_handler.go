package handler

import (
	"strconv"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type GambarKerjaHandler struct {
	svc service.GambarKerjaService
}

func NewGambarKerjaHandler(svc service.GambarKerjaService) *GambarKerjaHandler {
	return &GambarKerjaHandler{svc: svc}
}

func (h *GambarKerjaHandler) getUserEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return "CS / Designer / Estimator"
	}
	return email.(string)
}

func (h *GambarKerjaHandler) getUserID(c *gin.Context) uint {
	id, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return id.(uint)
}

// GET /api/gambar-kerja
func (h *GambarKerjaHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Data gambar kerja berhasil dimuat", data)
}

// POST /api/gambar-kerja/:id/response
// Di sini :id adalah Order ID
func (h *GambarKerjaHandler) Response(c *gin.Context) {
	orderID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Order ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.Response(c.Request.Context(), orderID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil response gambar kerja order", data)
}

// POST /api/gambar-kerja/upload
func (h *GambarKerjaHandler) Upload(c *gin.Context) {
	orderIDStr := c.PostForm("order_id")
	if orderIDStr == "" {
		helper.BadRequest(c, "order_id wajib disertakan", nil)
		return
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		helper.BadRequest(c, "order_id tidak valid", nil)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		helper.BadRequest(c, "Failed to parse multipart form", nil)
		return
	}

	files := form.File["gambar_kerja[]"]
	if len(files) == 0 {
		files = form.File["gambar_kerja"]
	}

	if len(files) == 0 {
		helper.BadRequest(c, "gambar_kerja[] atau gambar_kerja wajib diunggah", nil)
		return
	}

	userID := h.getUserID(c)
	data, err := h.svc.Upload(c.Request.Context(), uint(orderID), files, userID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Berhasil mengunggah gambar kerja", data)
}

// POST /api/gambar-kerja/files/:file_id/revise
func (h *GambarKerjaHandler) ReviseFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		helper.BadRequest(c, "File ID tidak valid", nil)
		return
	}

	var req dto.ReviseWorkingDrawingFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	err = h.svc.ReviseFile(c.Request.Context(), uint(fileID), req.Notes)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "File gambar kerja berhasil ditandai revisi", nil)
}

// POST /api/gambar-kerja/:id/revise-general
// Di sini :id adalah GambarKerja ID
func (h *GambarKerjaHandler) ReviseGeneral(c *gin.Context) {
	gkID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Gambar Kerja ID tidak valid", nil)
		return
	}

	var req dto.ReviseGeneralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.ReviseGeneral(c.Request.Context(), gkID, req.Notes, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Revisi general berhasil dikirim", data)
}

// POST /api/gambar-kerja/:id/approve
// Di sini :id adalah GambarKerja ID
func (h *GambarKerjaHandler) Approve(c *gin.Context) {
	gkID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Gambar Kerja ID tidak valid", nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.Approve(c.Request.Context(), gkID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Gambar kerja berhasil disetujui", data)
}

// DELETE /api/gambar-kerja/files/:file_id
func (h *GambarKerjaHandler) DeleteFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		helper.BadRequest(c, "File ID tidak valid", nil)
		return
	}

	err = h.svc.DeleteFile(c.Request.Context(), uint(fileID))
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "File gambar kerja berhasil dihapus", nil)
}
