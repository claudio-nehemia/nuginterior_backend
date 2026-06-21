package handler

import (
	"io"

	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	aiSvc        service.AIService
	dashboardSvc service.DashboardService
}

func NewAIHandler(aiSvc service.AIService, dashboardSvc service.DashboardService) *AIHandler {
	return &AIHandler{
		aiSvc:        aiSvc,
		dashboardSvc: dashboardSvc,
	}
}

func (h *AIHandler) GetSurveySummary(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	analysis, err := h.aiSvc.GenerateSurveyBrief(c.Request.Context(), id)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Brief desain AI berhasil dibuat", gin.H{"summary": analysis})
}

func (h *AIHandler) GetGlobalAnalysis(c *gin.Context) {
	// Fetch actual dashboard statistics first
	stats, err := h.dashboardSvc.GetStats(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat data statistik untuk analisis AI")
		return
	}

	analysis, err := h.aiSvc.AnalyzeGlobalDashboard(c.Request.Context(), *stats)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Analisis eksekutif dashboard AI berhasil dibuat", gin.H{"analysis": analysis})
}

func (h *AIHandler) GetProjectHealth(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	analysis, err := h.aiSvc.AnalyzeProjectHealth(c.Request.Context(), id)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Analisis kesehatan proyek AI berhasil dibuat", gin.H{"health_report": analysis})
}

func (h *AIHandler) Transcribe(c *gin.Context) {
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		helper.BadRequest(c, "Audio file is required", nil)
		return
	}
	defer file.Close()

	// Read audio file bytes
	audioBytes, err := io.ReadAll(file)
	if err != nil {
		helper.InternalError(c, "Failed to read audio file")
		return
	}

	transcript, err := h.aiSvc.TranscribeAudio(c.Request.Context(), audioBytes, header.Filename)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Transkripsi suara berhasil", gin.H{"transcript": transcript})
}
