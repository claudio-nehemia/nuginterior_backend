package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIService interface {
	GenerateSurveyBrief(ctx context.Context, surveyID uint) (string, error)
	AnalyzeGlobalDashboard(ctx context.Context, stats dto.DashboardStatsResponse) (string, error)
	AnalyzeProjectHealth(ctx context.Context, orderID uint) (string, error)
	TranscribeAudio(ctx context.Context, audioBytes []byte, filename string) (string, error)
}

type aiService struct {
	db        *gorm.DB
	openaiKey string
	logger    *zap.Logger
}

func NewAIService(db *gorm.DB, openaiKey string, logger *zap.Logger) AIService {
	return &aiService{
		db:        db,
		openaiKey: openaiKey,
		logger:    logger,
	}
}

// ChatMessage represents a single message in the ChatGPT conversation history.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the payload for the OpenAI Chat Completions API.
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// ChatResponse represents the OpenAI API response.
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *aiService) callGPT(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if s.openaiKey == "" {
		return "", errors.New("OpenAI API Key is not configured. Please add OPENAI_API_KEY to your .env file")
	}

	payload := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.openaiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decode response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if apiResp.Error != nil {
			return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, apiResp.Error.Message)
		}
		return "", fmt.Errorf("OpenAI API returned non-OK status: %d", resp.StatusCode)
	}

	if len(apiResp.Choices) == 0 {
		return "", errors.New("OpenAI API returned empty choices")
	}

	return apiResp.Choices[0].Message.Content, nil
}

func (s *aiService) GenerateSurveyBrief(ctx context.Context, surveyID uint) (string, error) {
	var survey entity.Survey
	err := s.db.WithContext(ctx).
		Preload("Order").
		Preload("Surveyor").
		Preload("SurveyPengukuran.JenisPengukuran").
		First(&survey, surveyID).Error
	if err != nil {
		return "", fmt.Errorf("survey not found: %w", err)
	}

	var notes string
	if survey.Catatan != "" {
		notes = survey.Catatan
	} else {
		notes = "Tidak ada catatan survey khusus."
	}

	measurements := ""
	for _, p := range survey.SurveyPengukuran {
		checkedStr := "No"
		if p.Checked {
			checkedStr = "Yes"
		}
		name := p.NamaCustom
		if p.JenisPengukuran != nil {
			name = p.JenisPengukuran.NamaPengukuran
		}
		measurements += fmt.Sprintf("- Pengukuran: %s | Checked: %s | Notes: %s | P: %.2f, L: %.2f, T: %.2f\n", 
			name, checkedStr, p.Notes, p.Panjang, p.Lebar, p.Tinggi)
	}

	projectName := "-"
	customerName := "-"
	interiorType := "-"
	if survey.Order != nil {
		projectName = survey.Order.NamaProject
		customerName = survey.Order.NamaCustomer
		interiorType = survey.Order.JenisInterior
	}

	systemPrompt := "You are a professional senior interior design assistant for Arsiflow. Your job is to read raw survey notes and measurements and generate a clean, detailed, and professional design brief (instructions) for the designer in Indonesian. Output your response in structured Markdown."

	userPrompt := fmt.Sprintf(`### DATA SURVEY PROYEK ARSIFLOW
- **Nama Proyek**: %s
- **Nama Customer**: %s
- **Jenis Interior**: %s
- **Lokasi Proyek**: %s
- **Catatan Surveyor**: %s

### DATA PENGUKURAN DAN TEMUAN LAPANGAN:
%s

Tolong buatkan ringkasan instruksi brief desain yang terstruktur dalam Bahasa Indonesia. Brief harus memuat:
1. **Analisis Ruangan & Karakter Proyek**: Kesimpulan singkat mengenai potensi dan tantangan ruangan.
2. **Kebutuhan Utama Desain (Must-Haves)**: Daftar item yang wajib didesain atau diperhatikan berdasarkan catatan survey.
3. **Rekomendasi Gaya & Layout**: Usulan penempatan furnitur/layout dan skema warna yang cocok dengan tipe interior.
4. **Catatan Teknis Penting**: Catatan dimensi/ukuran penting dari hasil pengukuran lapangan yang krusial bagi desainer (misal tinggi plafon, dsb).
Format output dalam Markdown yang terstruktur dan indah.`, projectName, customerName, interiorType, survey.Lokasi, notes, measurements)

	return s.callGPT(ctx, systemPrompt, userPrompt)
}

func (s *aiService) AnalyzeGlobalDashboard(ctx context.Context, stats dto.DashboardStatsResponse) (string, error) {
	systemPrompt := "You are a senior business analyst and executive assistant at Arsiflow (premium interior design & architecture firm). Your job is to analyze the company's overall statistics, omset, cashflow, and active projects, and provide a concise, high-level business report for the director. Output your response in structured Markdown in Indonesian."

	userPrompt := fmt.Sprintf(`### DATA UTAMA PERUSAHAAN (DASBOARD GLOBAL)
- **Total Seluruh Order**: %d
- **Order Aktif (Sedang Berjalan)**: %d
- **Proyek Selesai**: %d
- **Success Rate Proyek**: %.2f%%
- **Jumlah Kontrak Deal (Signed)**: %d
- **Finansial Lunas**: %d Proyek (Total: Rp%.2f)
- **Finansial Belum Terbayar (Outstanding Invoices)**: %d Proyek (Total Tagihan Belum Terbayar: Rp%.2f)
- **Total Omset Terdaftar (Keseluruhan Invoice)**: Rp%.2f

Tolong berikan laporan ringkas (Executive Summary) dalam Bahasa Indonesia mengenai:
1. **Analisis Kesehatan Finansial**: Apakah kas perusahaan sehat berdasarkan omset, tagihan lunas vs outstanding bill.
2. **Performa Proyek**: Keberhasilan penyelesaian proyek dan kapasitas tim saat ini (order aktif).
3. **Rekomendasi Strategis**: Tindakan segera apa yang harus diambil (misal mengejar tagihan outstanding, melakukan ekspansi marketing, dsb).
Gunakan nada profesional, ringkas, dan langsung pada intinya.`, 
		stats.TotalOrders, stats.ActiveOrders, stats.CompletedProjects, stats.SuccessRate, 
		stats.TotalContractsDeal, stats.LunasCount, stats.LunasAmount, 
		stats.BelumBayarCount, stats.BelumBayarAmount, stats.TotalOmset)

	return s.callGPT(ctx, systemPrompt, userPrompt)
}

func (s *aiService) AnalyzeProjectHealth(ctx context.Context, orderID uint) (string, error) {
	var order entity.Order
	err := s.db.WithContext(ctx).
		Preload("Contracts.RAB").
		Preload("Teams.User").
		Preload("PIC").
		First(&order, orderID).Error
	if err != nil {
		return "", fmt.Errorf("order not found: %w", err)
	}

	var invoices []entity.Invoice
	s.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&invoices)

	var workplan entity.Workplan
	hasWorkplan := true
	err = s.db.WithContext(ctx).
		Preload("Stages.StageMaster").
		Where("order_id = ?", orderID).
		First(&workplan).Error
	if err != nil {
		hasWorkplan = false
	}

	// Finance aggregation
	totalRAB := 0.0
	for _, c := range order.Contracts {
		if c.Status == "deal" && c.RAB != nil {
			totalRAB = c.RAB.GrandTotal
			break
		}
	}
	if totalRAB == 0 {
		totalRAB = order.HargaKontrak.InexactFloat64()
	}

	totalInvoiced := 0.0
	totalPaid := 0.0
	totalUnpaid := 0.0
	for _, inv := range invoices {
		totalInvoiced += inv.Amount
		if inv.Status == "terbayar" {
			totalPaid += inv.Amount
		} else {
			totalUnpaid += inv.Amount
		}
	}

	// Timeline aggregation
	stagesStr := ""
	delayedStages := 0
	completedStages := 0
	totalStages := 0
	if hasWorkplan {
		totalStages = len(workplan.Stages)
		now := time.Now()
		for _, stage := range workplan.Stages {
			name := ""
			if stage.StageMaster != nil {
				name = stage.StageMaster.Name
			}
			stagesStr += fmt.Sprintf("- Tahap: %s | Persentase: %.1f%% | Status: %s\n", name, stage.Percentage, stage.Status)
			if stage.Status == "completed" {
				completedStages++
			} else if stage.Status == "pending" && stage.EndDate != nil && stage.EndDate.Before(now) {
				delayedStages++
			}
		}
	}

	systemPrompt := "You are a senior project manager at Arsiflow. Your job is to audit a project's health using its financial and timeline data. You must provide a project health status (HEALTHY, WARNING, or CRITICAL) followed by a detailed audit report. Output your response in structured Markdown in Indonesian."

	workplanStatus := "Belum dibuat"
	if hasWorkplan {
		workplanStatus = fmt.Sprintf("Sudah dibuat (Total tahapan: %d, Selesai: %d, Terlambat: %d)", totalStages, completedStages, delayedStages)
	}

	userPrompt := fmt.Sprintf(`### DATA PROYEK
- **Nomor Order**: %s
- **Nama Proyek**: %s
- **Customer**: %s
- **Tahapan Proyek**: %s
- **Status Proyek**: %s
- **Status Pembayaran**: %s

### KEUANGAN PROYEK
- **Nilai Kontrak (RAB/Deal)**: Rp%.2f
- **Total Ditagihkan (Invoice)**: Rp%.2f
- **Total Terbayar**: Rp%.2f
- **Tagihan Belum Terbayar (Outstanding)**: Rp%.2f

### JADWAL & TIMELINE PROYEK
- **Status Workplan**: %s
- **Rincian Tahapan Kerja**:
%s

Tolong buatkan audit kesehatan proyek terstruktur dalam Bahasa Indonesia yang memuat:
1. **Status Kesehatan Proyek**: Tentukan status kesehatan (pilih salah satu dari: **HEALTHY** [jika semua aman], **WARNING** [ada keterlambatan minor/belum bayar termin], **CRITICAL** [keterlambatan parah / invoice tertunda lama]). Berikan visualisasi badge status yang mencolok di awal.
2. **Analisis Finansial**: Apakah pembayaran lancar dan sesuai termin pekerjaan.
3. **Analisis Progress & Jadwal**: Apakah pengerjaan di lapangan/workshop tepat waktu atau ada kemacetan tahap.
4. **Rekomendasi Tindakan PM**: Daftar tugas yang harus segera dilakukan oleh Project Manager hari ini.
Format dalam Markdown terstruktur dengan rapi.`, 
		order.NomorOrder, order.NamaProject, order.NamaCustomer, order.TahapanProyek, 
		order.ProjectStatus, order.PaymentStatus, totalRAB, totalInvoiced, totalPaid, totalUnpaid,
		workplanStatus, stagesStr)

	return s.callGPT(ctx, systemPrompt, userPrompt)
}

func (s *aiService) TranscribeAudio(ctx context.Context, audioBytes []byte, filename string) (string, error) {
	if s.openaiKey == "" {
		return "", errors.New("OpenAI API Key is not configured. Please add OPENAI_API_KEY to your .env file")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create the audio file form field
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file failed: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(audioBytes)); err != nil {
		return "", fmt.Errorf("copy audio bytes failed: %w", err)
	}

	// Create the model form field
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return "", fmt.Errorf("write model field failed: %w", err)
	}

	// Create optional language parameter to help Whisper with Indonesian transcription
	if err := writer.WriteField("language", "id"); err != nil {
		return "", fmt.Errorf("write language field failed: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+s.openaiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API returned non-OK status: %d, response: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decode response failed: %w", err)
	}

	return apiResp.Text, nil
}
