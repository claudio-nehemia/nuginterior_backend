package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SurveyHandler struct {
	svc        service.SurveyService
	settingSvc service.SettingService
}

func NewSurveyHandler(svc service.SurveyService, settingSvc service.SettingService) *SurveyHandler {
	return &SurveyHandler{svc: svc, settingSvc: settingSvc}
}

func (h *SurveyHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Data survey berhasil dimuat", data)
}

func (h *SurveyHandler) Store(c *gin.Context) {
	var req dto.CreateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	userIDVal, exists := c.Get("user_id")
	if exists {
		uid := userIDVal.(uint)
		req.SurveyorID = &uid
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.Created(c, "Survey berhasil dibuat", data)
}

func (h *SurveyHandler) Show(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	data, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		helper.NotFound(c, "Survey tidak ditemukan")
		return
	}
	helper.OK(c, "Detail survey", data)
}

func (h *SurveyHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.UpdateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	userIDVal, exists := c.Get("user_id")
	if exists {
		uid := userIDVal.(uint)
		req.SurveyorID = &uid
	}
	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Survey berhasil diupdate", data)
}

func (h *SurveyHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Survey berhasil dihapus", nil)
}

func (h *SurveyHandler) Response(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	enabled, _ := h.settingSvc.IsEnabled(c.Request.Context(), "response_enabled")
	if !enabled {
		helper.Forbidden(c, "Fitur response tidak aktif")
		return
	}
	email, _ := c.Get("email")
	data, err := h.svc.Response(c.Request.Context(), id, email.(string))
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Response berhasil", data)
}

func (h *SurveyHandler) MarketingResponse(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	enabled, _ := h.settingSvc.IsEnabled(c.Request.Context(), "marketing_response_enabled")
	if !enabled {
		helper.Forbidden(c, "Fitur marketing response tidak aktif")
		return
	}
	email, _ := c.Get("email")
	data, err := h.svc.MarketingResponse(c.Request.Context(), id, email.(string))
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Marketing response berhasil", data)
}
