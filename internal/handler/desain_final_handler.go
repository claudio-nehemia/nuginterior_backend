package handler

import (
	"strconv"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DesainFinalHandler struct {
	svc service.DesainFinalService
}

func NewDesainFinalHandler(svc service.DesainFinalService) *DesainFinalHandler {
	return &DesainFinalHandler{svc: svc}
}

func (h *DesainFinalHandler) getUserEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return "CS / Designer / Estimator"
	}
	return email.(string)
}

// GET /api/desain-finals
func (h *DesainFinalHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Data desain final berhasil dimuat", data)
}

// POST /api/desain-finals/:id/response
// Di sini :id adalah Order ID
func (h *DesainFinalHandler) Response(c *gin.Context) {
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
	helper.OK(c, "Berhasil response desain final order", data)
}

// POST /api/desain-finals/upload
func (h *DesainFinalHandler) Upload(c *gin.Context) {
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

	files := form.File["moodboard_final[]"]
	if len(files) == 0 {
		files = form.File["moodboard_final"]
	}

	if len(files) == 0 {
		helper.BadRequest(c, "moodboard_final[] atau moodboard_final wajib diunggah", nil)
		return
	}

	data, err := h.svc.Upload(c.Request.Context(), uint(orderID), files)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Berhasil mengunggah desain final", data)
}

// POST /api/desain-finals/:id/accept
// Di sini :id adalah DesainFinal ID
func (h *DesainFinalHandler) Accept(c *gin.Context) {
	desainFinalID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Desain Final ID tidak valid", nil)
		return
	}

	var req dto.AcceptDesainFinalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.Accept(c.Request.Context(), desainFinalID, req.DesainFinalFileID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Desain final berhasil disetujui sebagai master blueprint", data)
}

// POST /api/desain-finals/:id/revise
// Di sini :id adalah DesainFinal ID
func (h *DesainFinalHandler) Revise(c *gin.Context) {
	desainFinalID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Desain Final ID tidak valid", nil)
		return
	}

	var req dto.ReviseDesainFinalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.Revise(c.Request.Context(), desainFinalID, req.DesainFinalFileID, req.Notes, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Status revisi desain final berhasil disimpan", data)
}

// DELETE /api/desain-finals/files/:file_id
func (h *DesainFinalHandler) DeleteFile(c *gin.Context) {
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

	helper.OK(c, "File desain final berhasil dihapus", nil)
}
