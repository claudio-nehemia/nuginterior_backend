package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderService interface {
	GetAll(ctx context.Context, search, status string) ([]dto.OrderResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.OrderResponse, error)
	Create(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateOrderRequest) (*dto.OrderResponse, error)
	Delete(ctx context.Context, id uint) error
	SyncTeams(ctx context.Context, orderID uint, userIDs []uint) ([]dto.OrderTeamResponse, error)
	GetTeams(ctx context.Context, orderID uint) ([]dto.OrderTeamResponse, error)
	ExportPDF(ctx context.Context, params map[string]string) ([]byte, string, error)
	ExportExcel(ctx context.Context, params map[string]string) ([]byte, string, error)
}

type orderService struct {
	repo            repository.OrderRepository
	db              *gorm.DB
	logger          *zap.Logger
	uploadDir       string
	logTaskSvc      ProjectLogTaskService
	notificationSvc NotificationService
}

func NewOrderService(repo repository.OrderRepository, db *gorm.DB, logger *zap.Logger, uploadDir string, logTaskSvc ProjectLogTaskService, notificationSvc NotificationService) OrderService {
	return &orderService{
		repo:            repo,
		db:              db,
		logger:          logger,
		uploadDir:       uploadDir,
		logTaskSvc:      logTaskSvc,
		notificationSvc: notificationSvc,
	}
}

func (s *orderService) GetAll(ctx context.Context, search, status string) ([]dto.OrderResponse, error) {
	list, err := s.repo.FindAll(ctx, search, status)
	if err != nil {
		return nil, err
	}
	result := make([]dto.OrderResponse, len(list))
	for i, o := range list {
		result[i] = toOrderResponse(o, false)
	}
	return result, nil
}

func (s *orderService) GetByID(ctx context.Context, id uint) (*dto.OrderResponse, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toOrderResponse(*o, true)
	return &resp, nil
}

func (s *orderService) Create(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	nomor, err := s.repo.GenerateNomorOrder(ctx)
	if err != nil {
		return nil, err
	}
	order := &entity.Order{
		NomorOrder:             nomor,
		NamaProject:            req.NamaProject,
		JenisInterior:          req.JenisInterior,
		NamaCustomer:           req.NamaCustomer,
		TeleponCustomer:        req.TeleponCustomer,
		EmailCustomer:          req.EmailCustomer,
		NamaPerusahaan:         req.NamaPerusahaan,
		CustomerAdditionalInfo: req.CustomerAdditionalInfo,
		NomorUnit:              req.NomorUnit,
		Alamat:                 req.Alamat,
		Catatan:                req.Catatan,
		ProjectStatus:          "pending",
		TahapanProyek:          "survey",
		TanggalSurvey:          req.TanggalSurvey,
	}

	if req.TanggalMasukCustomer != "" {
		t, err := time.Parse("2006-01-02", req.TanggalMasukCustomer)
		if err == nil {
			order.TanggalMasukCustomer = &t
		}
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	_ = s.logTaskSvc.TransitionStage(ctx, order.ID, "survey", "")

	resp := toOrderResponse(*order, false)
	return &resp, nil
}

func (s *orderService) Update(ctx context.Context, id uint, req dto.UpdateOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	order.NamaProject = req.NamaProject
	order.JenisInterior = req.JenisInterior
	order.NamaCustomer = req.NamaCustomer
	order.TeleponCustomer = req.TeleponCustomer
	order.EmailCustomer = req.EmailCustomer
	order.NamaPerusahaan = req.NamaPerusahaan
	order.CustomerAdditionalInfo = req.CustomerAdditionalInfo
	order.NomorUnit = req.NomorUnit
	order.Alamat = req.Alamat
	order.Catatan = req.Catatan
	order.TanggalSurvey = req.TanggalSurvey

	if req.ProjectStatus != "" {
		order.ProjectStatus = req.ProjectStatus
	}

	if req.PriorityLevel != "" {
		order.PriorityLevel = req.PriorityLevel
	}

	if req.TahapanProyek != "" {
		oldStage := order.TahapanProyek
		if oldStage != req.TahapanProyek {
			_ = s.logTaskSvc.TransitionStage(ctx, order.ID, req.TahapanProyek, "")
		}
	}

	if req.TanggalMasukCustomer != "" {
		t, _ := time.Parse("2006-01-02", req.TanggalMasukCustomer)
		order.TanggalMasukCustomer = &t
	}

	if req.TanggalMulai != "" {
		t, _ := time.Parse("2006-01-02", req.TanggalMulai)
		order.TanggalMulai = &t
	}

	if req.TanggalSelesai != "" {
		t, _ := time.Parse("2006-01-02", req.TanggalSelesai)
		order.TanggalSelesai = &t
	}

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	resp := toOrderResponse(*order, false)
	return &resp, nil
}

func (s *orderService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *orderService) SyncTeams(ctx context.Context, orderID uint, userIDs []uint) ([]dto.OrderTeamResponse, error) {
	if err := s.repo.SyncTeams(ctx, orderID, userIDs); err != nil {
		return nil, err
	}

	// Trigger assign_order notification
	order, err := s.repo.FindByID(ctx, orderID)
	if err == nil && order != nil {
		_ = s.notificationSvc.SendNotification(
			ctx,
			orderID,
			"assign_order",
			"Tim Proyek Ditugaskan",
			fmt.Sprintf("Anda telah ditugaskan ke dalam tim proyek untuk %s (Order %s).", order.NamaProject, order.NomorOrder),
			fmt.Sprintf("/dashboard/order/%d", orderID),
		)
	}

	return s.GetTeams(ctx, orderID)
}

func (s *orderService) GetTeams(ctx context.Context, orderID uint) ([]dto.OrderTeamResponse, error) {
	teams, err := s.repo.GetTeams(ctx, orderID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.OrderTeamResponse, len(teams))
	for i, t := range teams {
		result[i] = dto.OrderTeamResponse{
			ID:     t.ID,
			UserID: t.UserID,
		}
		if t.User != nil {
			result[i].Name = t.User.Name
			result[i].Email = t.User.Email
			if t.User.Role != nil {
				result[i].Role = t.User.Role.NamaRole
			}
		}
	}
	return result, nil
}

func toOrderResponse(o entity.Order, includeDetails bool) dto.OrderResponse {
	resp := dto.OrderResponse{
		ID:                     o.ID,
		NomorOrder:             o.NomorOrder,
		NamaProject:            o.NamaProject,
		JenisInterior:          o.JenisInterior,
		NamaCustomer:           o.NamaCustomer,
		TeleponCustomer:        o.TeleponCustomer,
		EmailCustomer:          o.EmailCustomer,
		NamaPerusahaan:         o.NamaPerusahaan,
		CustomerAdditionalInfo: o.CustomerAdditionalInfo,
		NomorUnit:              o.NomorUnit,
		Alamat:                 o.Alamat,
		Catatan:                o.Catatan,
		TanggalMasukCustomer:   o.TanggalMasukCustomer,
		ProjectStatus:          o.ProjectStatus,
		PriorityLevel:          o.PriorityLevel,
		TahapanProyek:          o.TahapanProyek,
		PaymentStatus:          o.PaymentStatus,
		TerminID:               o.TerminID,
		HargaKontrak:           o.HargaKontrak.String(),
		TanggalKontrak:         o.TanggalKontrak,
		NomorKontrak:           o.NomorKontrak,
		TanggalMulai:           o.TanggalMulai,
		TanggalSelesai:         o.TanggalSelesai,
		PicID:                  o.PicID,
		MomFile:                o.MomFile,
		MomFiles:               json.RawMessage(o.MomFiles),
		TanggalSurvey:          o.TanggalSurvey,
		SurveyResponseBy:       o.SurveyResponseBy,
		SurveyResponseTime:     o.SurveyResponseTime,
		PmSurveyResponseBy:     o.PmSurveyResponseBy,
		PmSurveyResponseTime:   o.PmSurveyResponseTime,
		CreatedAt:              o.CreatedAt,
		UpdatedAt:              o.UpdatedAt,
	}

	// Always map teams if loaded
	if len(o.Teams) > 0 {
		resp.Teams = make([]dto.OrderTeamResponse, len(o.Teams))
		for i, t := range o.Teams {
			resp.Teams[i] = dto.OrderTeamResponse{
				ID:     t.ID,
				UserID: t.UserID,
			}
			if t.User != nil {
				resp.Teams[i].Name = t.User.Name
				resp.Teams[i].Email = t.User.Email
				if t.User.Role != nil {
					resp.Teams[i].Role = t.User.Role.NamaRole
				}
			}
		}
	}

	if includeDetails {
		if len(o.Surveys) > 0 {
			resp.Surveys = make([]dto.SurveyResponse, len(o.Surveys))
			for i, s := range o.Surveys {
				resp.Surveys[i] = toSurveyResponse(s, false)
			}
		}
		if len(o.Moodboards) > 0 {
			resp.Moodboards = make([]dto.MoodboardResponse, len(o.Moodboards))
			for i, m := range o.Moodboards {
				resp.Moodboards[i] = toMoodboardResponse(m, false)
			}
		}
	}

	return resp
}

func toSurveyResponse(s entity.Survey, includeOrder bool) dto.SurveyResponse {
	resp := dto.SurveyResponse{
		ID:                    s.ID,
		OrderID:               s.OrderID,
		TanggalSurvey:         s.TanggalSurvey,
		Lokasi:                s.Lokasi,
		Catatan:               s.Catatan,
		Status:                s.Status,
		SurveyorID:            s.SurveyorID,
		ResponseBy:            s.ResponseBy,
		ResponseTime:          s.ResponseTime,
		MarketingResponseBy:   s.MarketingResponseBy,
		MarketingResponseTime: s.MarketingResponseTime,
		CreatedAt:             s.CreatedAt,
		UpdatedAt:             s.UpdatedAt,
	}

	if includeOrder && s.Order != nil {
		resp.Order = &dto.OrderBriefResponse{
			ID:                   s.Order.ID,
			NomorOrder:           s.Order.NomorOrder,
			NamaProject:          s.Order.NamaProject,
			NamaCustomer:         s.Order.NamaCustomer,
			NamaPerusahaan:       s.Order.NamaPerusahaan,
			JenisInterior:        s.Order.JenisInterior,
			TanggalMasukCustomer: s.Order.TanggalMasukCustomer,
		}
	}

	if len(s.SurveyPengukuran) > 0 {
		resp.Pengukuran = make([]dto.PengukuranResponse, len(s.SurveyPengukuran))
		for i, p := range s.SurveyPengukuran {
			resp.Pengukuran[i] = dto.PengukuranResponse{
				ID:                p.ID,
				JenisPengukuranID: p.JenisPengukuranID,
				NamaPengukuran:    getPengukuranName(p.JenisPengukuran),
				Checked:           p.Checked,
				Notes:             p.Notes,
			}
		}
	}

	return resp
}

func toMoodboardResponse(m entity.Moodboard, includeOrder bool) dto.MoodboardResponse {
	return toMoodboardResponseEntity(m, includeOrder)
}

func getPengukuranName(jp *entity.JenisPengukuran) string {
	if jp == nil {
		return ""
	}
	return jp.NamaPengukuran
}

func (s *orderService) ExportPDF(ctx context.Context, params map[string]string) ([]byte, string, error) {
	orders, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, "", err
	}

	// Initialize A4 Landscape PDF
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Get Company Profile
	companyID, _ := ctx.Value(constants.ContextKeyCompanyID).(uint)
	if companyID == 0 {
		companyID = 1
	}
	cp := entity.GetCompanyProfile(s.db, companyID)

	// Draw Company Header
	logoFile := ""
	if cp.Logo != "" {
		logoFile = filepath.Join(s.uploadDir, filepath.Base(cp.Logo))
		if _, err := os.Stat(logoFile); os.IsNotExist(err) {
			logoFile = ""
		}
	}

	if logoFile != "" {
		pdf.Image(logoFile, 15, 12, 0, 15, false, "", 0, "")
		pdf.SetLeftMargin(33)
		pdf.SetX(33)
		pdf.SetY(12)
	} else {
		pdf.SetY(12)
	}

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 128, 128) // Teal
	pdf.CellFormat(0, 6, cp.Name, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 4, "Premium Interior Design & Architecture Services", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, fmt.Sprintf("Email: %s | Phone: %s | Alamat: %s", cp.Email, cp.Phone, cp.Address), "", 1, "L", false, 0, "")

	// Reset margins
	pdf.SetLeftMargin(15)
	pdf.SetX(15)
	pdf.SetY(29)

	pdf.SetDrawColor(0, 128, 128)
	pdf.SetLineWidth(0.8)
	pdf.Line(15, pdf.GetY(), 282, pdf.GetY())
	pdf.Ln(4)

	// Title
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 6, "LAPORAN DATA ORDER PROYEK NUGINTERIOR", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 4, fmt.Sprintf("Di-export pada: %s", time.Now().Format("02 January 2006 15:04:05")), "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Table Header
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(0, 128, 128)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.2)

	cols := []struct {
		name  string
		width float64
		align string
	}{
		{"No", 8, "C"},
		{"Nomor Order", 30, "L"},
		{"Nama Project", 38, "L"},
		{"Customer", 35, "L"},
		{"Jenis", 22, "L"},
		{"Status", 20, "C"},
		{"Tahapan", 22, "C"},
		{"Payment", 22, "C"},
		{"Priority", 16, "C"},
		{"Nilai Kontrak", 30, "R"},
		{"Tanggal Kontrak", 24, "C"},
	}

	for _, col := range cols {
		pdf.CellFormat(col.width, 7, col.name, "1", 0, col.align, true, 0, "")
	}
	pdf.Ln(7)

	// Table Rows
	pdf.SetFont("Arial", "", 7.5)
	pdf.SetTextColor(60, 60, 60)
	var totalContract float64
	for idx, o := range orders {
		// Alternate row colors
		if idx%2 == 0 {
			pdf.SetFillColor(255, 255, 255)
		} else {
			pdf.SetFillColor(245, 247, 248)
		}

		var hargaVal float64
		tglKontrakStr := "-"
		var dealContract *entity.Contract
		for _, c := range o.Contracts {
			if c.Status == "deal" {
				dealContract = &c
				break
			}
		}

		if dealContract != nil {
			if dealContract.RAB != nil {
				hargaVal = dealContract.RAB.GrandTotal
			}
			if dealContract.ResponseTime != nil {
				tglKontrakStr = dealContract.ResponseTime.Format("02/01/2006")
			} else {
				tglKontrakStr = dealContract.UpdatedAt.Format("02/01/2006")
			}
		} else {
			hargaVal = o.HargaKontrak.InexactFloat64()
			if o.TanggalKontrak != nil {
				tglKontrakStr = o.TanggalKontrak.Format("02/01/2006")
			}
		}

		totalContract += hargaVal

		// Helper to format string within width limit
		truncate := func(s string, w float64) string {
			if pdf.GetStringWidth(s) > w-2 {
				for len(s) > 0 && pdf.GetStringWidth(s+"...") > w-2 {
					s = s[:len(s)-1]
				}
				return s + "..."
			}
			return s
		}

		pdf.CellFormat(8, 6, fmt.Sprintf("%d", idx+1), "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 6, truncate(o.NomorOrder, 30), "1", 0, "L", true, 0, "")
		pdf.CellFormat(38, 6, truncate(o.NamaProject, 38), "1", 0, "L", true, 0, "")
		pdf.CellFormat(35, 6, truncate(o.NamaCustomer, 35), "1", 0, "L", true, 0, "")
		pdf.CellFormat(22, 6, truncate(o.JenisInterior, 22), "1", 0, "L", true, 0, "")
		pdf.CellFormat(20, 6, truncate(strings.ToUpper(o.ProjectStatus), 20), "1", 0, "C", true, 0, "")
		pdf.CellFormat(22, 6, truncate(strings.ToUpper(o.TahapanProyek), 22), "1", 0, "C", true, 0, "")
		pdf.CellFormat(22, 6, truncate(strings.ToUpper(o.PaymentStatus), 22), "1", 0, "C", true, 0, "")
		pdf.CellFormat(16, 6, truncate(strings.ToUpper(o.PriorityLevel), 16), "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 6, formatRupiah(hargaVal), "1", 0, "R", true, 0, "")
		pdf.CellFormat(24, 6, tglKontrakStr, "1", 1, "C", true, 0, "")
	}

	// Table Footer / Summary Row
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(230, 240, 240)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(213, 7, fmt.Sprintf("TOTAL (%d Order)", len(orders)), "1", 0, "R", true, 0, "")
	pdf.CellFormat(30, 7, formatRupiah(totalContract), "1", 0, "R", true, 0, "")
	pdf.CellFormat(24, 7, "", "1", 1, "C", true, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Laporan_Order_%s.pdf", time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

func (s *orderService) ExportExcel(ctx context.Context, params map[string]string) ([]byte, string, error) {
	orders, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	sheet := "Orders"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, "", err
	}
	f.SetActiveSheet(index)
	_ = f.DeleteSheet("Sheet1")

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"008080"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	// Set row style for data
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	currencyStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
		CustomNumFmt: ptr("\"Rp\"#,##0"), // Custom Rupiah / currency representation in Excel
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E6F0F0"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	// Set columns width
	colsWidth := map[string]float64{
		"A": 5,   // No
		"B": 20,  // Nomor Order
		"C": 30,  // Nama Project
		"D": 25,  // Nama Customer
		"E": 20,  // Jenis Interior
		"F": 15,  // Status Proyek
		"G": 18,  // Tahapan Proyek
		"H": 18,  // Payment Status
		"I": 15,  // Priority Level
		"J": 22,  // Harga Kontrak
		"K": 18,  // Tanggal Kontrak
	}
	for col, width := range colsWidth {
		_ = f.SetColWidth(sheet, col, col, width)
	}

	headers := []string{"No", "Nomor Order", "Nama Project", "Nama Customer", "Jenis Interior", "Status Proyek", "Tahapan Proyek", "Payment Status", "Priority Level", "Harga Kontrak", "Tanggal Kontrak"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	_ = f.SetRowHeight(sheet, 1, 25)

	rowIdx := 2
	for idx, o := range orders {
		var hargaVal float64
		tglKontrakStr := "-"
		var dealContract *entity.Contract
		for _, c := range o.Contracts {
			if c.Status == "deal" {
				dealContract = &c
				break
			}
		}

		if dealContract != nil {
			if dealContract.RAB != nil {
				hargaVal = dealContract.RAB.GrandTotal
			}
			if dealContract.ResponseTime != nil {
				tglKontrakStr = dealContract.ResponseTime.Format("02-01-2006")
			} else {
				tglKontrakStr = dealContract.UpdatedAt.Format("02-01-2006")
			}
		} else {
			hargaVal = o.HargaKontrak.InexactFloat64()
			if o.TanggalKontrak != nil {
				tglKontrakStr = o.TanggalKontrak.Format("02-01-2006")
			}
		}

		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), idx+1)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIdx), o.NomorOrder)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIdx), o.NamaProject)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", rowIdx), o.NamaCustomer)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", rowIdx), o.JenisInterior)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", rowIdx), strings.ToUpper(o.ProjectStatus))
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", rowIdx), strings.ToUpper(o.TahapanProyek))
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", rowIdx), strings.ToUpper(o.PaymentStatus))
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", rowIdx), strings.ToUpper(o.PriorityLevel))
		_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", rowIdx), hargaVal)
		_ = f.SetCellValue(sheet, fmt.Sprintf("K%d", rowIdx), tglKontrakStr)

		// Apply cell styles
		for i := 0; i < len(headers); i++ {
			cell := fmt.Sprintf("%c%d", 'A'+i, rowIdx)
			if i == 9 { // Harga Kontrak
				_ = f.SetCellStyle(sheet, cell, cell, currencyStyle)
			} else {
				_ = f.SetCellStyle(sheet, cell, cell, dataStyle)
			}
		}
		_ = f.SetRowHeight(sheet, rowIdx, 20)
		rowIdx++
	}

	// Add summary row
	_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), "TOTAL")
	_ = f.MergeCell(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("I%d", rowIdx))
	_ = f.SetCellFormula(sheet, fmt.Sprintf("J%d", rowIdx), fmt.Sprintf("=SUM(J2:J%d)", rowIdx-1))

	// Apply styles for summary row
	for i := 0; i < len(headers); i++ {
		cell := fmt.Sprintf("%c%d", 'A'+i, rowIdx)
		_ = f.SetCellStyle(sheet, cell, cell, summaryStyle)
		if i == 9 {
			// Maintain alignment and num format
			_ = f.SetCellStyle(sheet, cell, cell, currencyStyle)
			_ = f.SetCellStyle(sheet, cell, cell, summaryStyle) // Combined
		}
	}
	_ = f.SetRowHeight(sheet, rowIdx, 22)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Laporan_Order_%s.xlsx", time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

func ptr(s string) *string {
	return &s
}
