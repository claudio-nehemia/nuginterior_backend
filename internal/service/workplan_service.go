package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WorkplanService interface {
	GetAll(ctx context.Context) ([]dto.WorkplanResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.WorkplanResponse, error)
	GetByOrderID(ctx context.Context, orderID uint) (*dto.WorkplanResponse, error)
	Response(ctx context.Context, orderID uint, email string) (*dto.WorkplanResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateWorkplanRequest, email string) (*dto.WorkplanResponse, error)
	ExportExcel(ctx context.Context, id uint) ([]byte, string, error)
	GetStageMasters(ctx context.Context) ([]dto.WorkplanStageMasterResponse, error)
	UpdateStageMasters(ctx context.Context, req []dto.UpdateWorkplanStageMasterRequest) error
	CompleteStage(ctx context.Context, stageID uint, req dto.CompleteStageRequest, userEmail string) error
	RequestExtension(ctx context.Context, wpID uint, req dto.RequestExtensionRequest) (*dto.WorkplanResponse, error)
	HandleExtension(ctx context.Context, wpID uint, req dto.HandleExtensionRequest) (*dto.WorkplanResponse, error)
	ExportProgressPDF(ctx context.Context, id uint) ([]byte, string, error)
	ExportProgressExcel(ctx context.Context, id uint) ([]byte, string, error)
	// Defect Management
	ReportDefect(ctx context.Context, stageID uint, req dto.ReportDefectRequest, userEmail string) (*dto.WorkplanDefectResponse, error)
	SubmitDefectFix(ctx context.Context, defectID uint, req dto.SubmitDefectFixRequest, userEmail string) (*dto.WorkplanDefectResponse, error)
	ReviewDefect(ctx context.Context, defectID uint, req dto.ReviewDefectRequest, userEmail string) (*dto.WorkplanDefectResponse, error)
	GetDefectsByWorkplan(ctx context.Context, wpID uint) ([]dto.WorkplanDefectResponse, error)
	// BAST Management
	SubmitBast(ctx context.Context, id uint, req dto.SubmitBastRequest, email string) (*dto.WorkplanResponse, error)
	GenerateBastPDF(ctx context.Context, id uint) ([]byte, string, error)
}

type workplanService struct {
	repo       repository.WorkplanRepository
	settingSvc SettingService
	db         *gorm.DB
	logger     *zap.Logger
	uploadDir  string
	logTaskSvc ProjectLogTaskService
}

func NewWorkplanService(
	repo repository.WorkplanRepository,
	settingSvc SettingService,
	db *gorm.DB,
	logger *zap.Logger,
	uploadDir string,
	logTaskSvc ProjectLogTaskService,
) WorkplanService {
	return &workplanService{
		repo:       repo,
		settingSvc: settingSvc,
		db:         db,
		logger:     logger,
		uploadDir:  uploadDir,
		logTaskSvc: logTaskSvc,
	}
}

func (s *workplanService) checkApprovalMaterialCompleted(ctx context.Context, orderID uint) error {
	var am entity.ApprovalMaterial
	err := s.db.WithContext(ctx).Where("order_id = ?", orderID).First(&am).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("workplan terkunci! Approval material belum diisi/selesai")
		}
		return err
	}

	if am.Status != "completed" {
		return errors.New("workplan terkunci! Status approval material belum disetujui/selesai")
	}

	return nil
}

func (s *workplanService) GetAll(ctx context.Context) ([]dto.WorkplanResponse, error) {
	// Query orders that have a completed ApprovalMaterial
	var eligibleOrders []entity.Order
	err := s.db.WithContext(ctx).
		Preload("PIC").
		Preload("Contracts").
		Joins("JOIN approval_materials am ON am.order_id = orders.id").
		Where("am.status = ?", "completed").
		Find(&eligibleOrders).Error
	if err != nil {
		return nil, err
	}

	var result []dto.WorkplanResponse
	for _, order := range eligibleOrders {
		wp, err := s.repo.FindByOrderID(ctx, order.ID)
		orderCopy := order
		brief := toOrderBriefResponse(&orderCopy)
		// Load PIC (Project Manager from Survey Ulang Team)
		var survey entity.Survey
		if errSurvey := s.db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id DESC").First(&survey).Error; errSurvey == nil {
			var teamIDs []uint
			if len(survey.SurveyUlangTeamIDs) > 0 {
				_ = json.Unmarshal(survey.SurveyUlangTeamIDs, &teamIDs)
			}
			if len(teamIDs) > 0 {
				var users []entity.User
				if errUsers := s.db.WithContext(ctx).Preload("Role").Where("id IN ?", teamIDs).Find(&users).Error; errUsers == nil {
					for _, u := range users {
						if u.Role != nil && (u.Role.NamaRole == "Project Manager" || u.Role.NamaRole == "Supervisor") {
							brief.PicName = u.Name
							break
						}
					}
					if brief.PicName == "" && len(users) > 0 {
						brief.PicName = users[0].Name
					}
				}
			}
		}
		if brief.PicName == "" && order.PIC != nil {
			brief.PicName = order.PIC.Name
		}

		// Count the number of products (input item rooms) for this order
		var productCount int64
		s.db.WithContext(ctx).
			Table("input_item_rooms").
			Joins("JOIN input_items ON input_items.id = input_item_rooms.input_item_id").
			Where("input_items.order_id = ?", order.ID).
			Count(&productCount)

		brief.ProductCount = int(productCount)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Not response'd/created yet
				result = append(result, dto.WorkplanResponse{
					OrderID: order.ID,
					Status:  "belum_isi", // virtual status for list view
					Order:   brief,
					Stages:  []dto.WorkplanStageResponse{},
				})
				continue
			}
			return nil, err
		}

		resp := s.toWorkplanResponse(*wp)
		resp.Order = brief
		// status from draft -> sebagian, submitted -> lengkap
		if resp.Status == "draft" {
			resp.Status = "sebagian"
		} else if resp.Status == "submitted" {
			resp.Status = "lengkap"
		}
		result = append(result, *resp)
	}

	return result, nil
}

func (s *workplanService) GetByID(ctx context.Context, id uint) (*dto.WorkplanResponse, error) {
	wp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := s.toWorkplanResponse(*wp)
	
	// Load project & PIC metadata
	var order entity.Order
	if errOrder := s.db.WithContext(ctx).Preload("PIC").Preload("Contracts").First(&order, wp.OrderID).Error; errOrder == nil {
		brief := toOrderBriefResponse(&order)
		// Load PIC (Project Manager from Survey Ulang Team)
		var survey entity.Survey
		if errSurvey := s.db.WithContext(ctx).Where("order_id = ?", order.ID).Order("id DESC").First(&survey).Error; errSurvey == nil {
			var teamIDs []uint
			if len(survey.SurveyUlangTeamIDs) > 0 {
				_ = json.Unmarshal(survey.SurveyUlangTeamIDs, &teamIDs)
			}
			if len(teamIDs) > 0 {
				var users []entity.User
				if errUsers := s.db.WithContext(ctx).Preload("Role").Where("id IN ?", teamIDs).Find(&users).Error; errUsers == nil {
					for _, u := range users {
						if u.Role != nil && (u.Role.NamaRole == "Project Manager" || u.Role.NamaRole == "Supervisor") {
							brief.PicName = u.Name
							break
						}
					}
					if brief.PicName == "" && len(users) > 0 {
						brief.PicName = users[0].Name
					}
				}
			}
		}
		if brief.PicName == "" && order.PIC != nil {
			brief.PicName = order.PIC.Name
		}
		resp.Order = brief
	}

	return resp, nil
}

func (s *workplanService) GetByOrderID(ctx context.Context, orderID uint) (*dto.WorkplanResponse, error) {
	if err := s.checkApprovalMaterialCompleted(ctx, orderID); err != nil {
		return nil, err
	}

	wp, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Check setting response_enabled
			isEnabled, _ := s.settingSvc.IsEnabled(ctx, "response_enabled")
			if !isEnabled {
				// Auto-initialize
				newWp, errCreate := s.initWorkplan(ctx, orderID, "System Auto")
				if errCreate != nil {
					return nil, errCreate
				}
				return s.GetByID(ctx, newWp.ID)
			}
			return nil, errors.New("workplan belum diresponse")
		}
		return nil, err
	}

	return s.GetByID(ctx, wp.ID)
}

func (s *workplanService) Response(ctx context.Context, orderID uint, email string) (*dto.WorkplanResponse, error) {
	if err := s.checkApprovalMaterialCompleted(ctx, orderID); err != nil {
		return nil, err
	}

	wp, err := s.repo.FindByOrderID(ctx, orderID)
	now := time.Now()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newWp, errCreate := s.initWorkplan(ctx, orderID, email)
			if errCreate != nil {
				return nil, errCreate
			}
			wp = newWp
		} else {
			return nil, err
		}
	} else {
		if wp.ResponseBy == "" {
			wp.ResponseBy = email
			wp.ResponseTime = &now
			if err := s.repo.Update(ctx, wp); err != nil {
				return nil, err
			}
		}
	}

	// Update order stage to workplan and log transition
	if err := s.logTaskSvc.TransitionStage(ctx, orderID, "workplan", email); err != nil {
		s.logger.Error("Failed to update order stage to workplan", zap.Error(err))
	}

	return s.GetByID(ctx, wp.ID)
}

func (s *workplanService) initWorkplan(ctx context.Context, orderID uint, userEmail string) (*entity.Workplan, error) {
	now := time.Now()
	newWp := &entity.Workplan{
		OrderID:      orderID,
		Status:       "draft",
		ResponseBy:   userEmail,
		ResponseTime: &now,
	}

	// Save main record first
	if err := s.repo.Create(ctx, newWp); err != nil {
		return nil, err
	}

	// Find the InputItem and rooms (products)
	var inputItem entity.InputItem
	err := s.db.WithContext(ctx).
		Preload("Rooms.Produk").
		Where("order_id = ?", orderID).
		First(&inputItem).Error
	if err != nil {
		s.logger.Warn("Failed to find InputItem during workplan init", zap.Error(err))
		return newWp, nil
	}

	// Get all Stage masters
	masters, err := s.repo.GetStageMasters(ctx)
	if err != nil {
		s.logger.Error("Failed to get workplan stage masters", zap.Error(err))
		return newWp, nil
	}

	// Create a default stage timeline for each room product & each stage master
	for _, room := range inputItem.Rooms {
		if room.ProdukID == nil {
			continue
		}
		for _, m := range masters {
			stage := &entity.WorkplanStage{
				WorkplanID:      newWp.ID,
				InputItemRoomID: room.ID,
				StageMasterID:   m.ID,
				Percentage:      m.Percentage,
				StartDate:       nil,
				EndDate:         nil,
				Notes:           "",
			}
			if err := s.repo.SaveStage(ctx, stage); err != nil {
				s.logger.Error("Failed to save default stage record", zap.Error(err))
			}
		}
	}

	return newWp, nil
}

func (s *workplanService) Update(ctx context.Context, id uint, req dto.UpdateWorkplanRequest, email string) (*dto.WorkplanResponse, error) {
	wp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if wp.Status == "submitted" {
		return nil, errors.New("workplan sudah disubmit dan tidak dapat diubah lagi")
	}

	// Fetch Order contract date ranges
	var order entity.Order
	if errOrder := s.db.WithContext(ctx).First(&order, wp.OrderID).Error; errOrder != nil {
		return nil, fmt.Errorf("failed to fetch associated order details: %w", errOrder)
	}

	// Keep track of stage dates for calculating min/max project dates
	var minDate, maxDate *time.Time

	// Perform all updates inside a transaction
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete all existing stages for this workplan to sync with the new request payload
		if errDel := tx.Where("workplan_id = ?", wp.ID).Delete(&entity.WorkplanStage{}).Error; errDel != nil {
			return fmt.Errorf("failed to clear old stages: %w", errDel)
		}

		// Re-insert new stage list from request
		for _, reqStage := range req.Stages {
			var stgStartDate, stgEndDate *time.Time
			if reqStage.StartDate != nil && *reqStage.StartDate != "" {
				parsed, errParse := time.Parse("2006-01-02", *reqStage.StartDate)
				if errParse != nil {
					return fmt.Errorf("format start_date salah pada stage ID: %d", reqStage.StageMasterID)
				}
				stgStartDate = &parsed

				if minDate == nil || parsed.Before(*minDate) {
					minDate = &parsed
				}
			}

			if reqStage.EndDate != nil && *reqStage.EndDate != "" {
				parsed, errParse := time.Parse("2006-01-02", *reqStage.EndDate)
				if errParse != nil {
					return fmt.Errorf("format end_date salah pada stage ID: %d", reqStage.StageMasterID)
				}
				stgEndDate = &parsed

				if maxDate == nil || parsed.After(*maxDate) {
					maxDate = &parsed
				}
			}

			// Validate stage range
			if stgStartDate != nil && stgEndDate != nil {
				if stgEndDate.Before(*stgStartDate) {
					return fmt.Errorf("tanggal selesai mendahului tanggal mulai pada stage")
				}
			}

			stageRecord := &entity.WorkplanStage{
				WorkplanID:      wp.ID,
				InputItemRoomID: reqStage.InputItemRoomID,
				StageMasterID:   reqStage.StageMasterID,
				Percentage:      reqStage.Percentage,
				StartDate:       stgStartDate,
				EndDate:         stgEndDate,
				Notes:           reqStage.Notes,
			}

			if errSave := tx.Create(stageRecord).Error; errSave != nil {
				return fmt.Errorf("failed to save stage record: %w", errSave)
			}
		}

		// Configure Overall Project dates
		var wpStartDate, wpEndDate *time.Time
		var wpDurationDays int

		if req.StartDate != nil && *req.StartDate != "" {
			parsed, errParse := time.Parse("2006-01-02", *req.StartDate)
			if errParse != nil {
				return errors.New("format start_date project salah")
			}
			wpStartDate = &parsed
		} else if minDate != nil {
			wpStartDate = minDate
		}

		if req.EndDate != nil && *req.EndDate != "" {
			parsed, errParse := time.Parse("2006-01-02", *req.EndDate)
			if errParse != nil {
				return errors.New("format end_date project salah")
			}
			wpEndDate = &parsed
		} else if maxDate != nil {
			wpEndDate = maxDate
		}

		if wpStartDate != nil && wpEndDate != nil {
			if wpEndDate.Before(*wpStartDate) {
				return errors.New("tanggal selesai proyek mendahului tanggal mulai proyek")
			}
			wpDurationDays = int(wpEndDate.Sub(*wpStartDate).Hours()/24) + 1

			// Validate against order contract duration
			if order.TanggalMulai != nil && order.TanggalSelesai != nil {
				maxContractDays := int(order.TanggalSelesai.Sub(*order.TanggalMulai).Hours()/24) + 1
				if wpDurationDays > maxContractDays {
					return fmt.Errorf("durasi workplan (%d hari) melebihi batas durasi kontrak proyek (%d hari)", wpDurationDays, maxContractDays)
				}
			}
		}

		wp.StartDate = wpStartDate
		wp.EndDate = wpEndDate
		wp.DurationDays = wpDurationDays

		// Set status transitions
		if req.Status == "submitted" {
			now := time.Now()
			wp.Status = "submitted"
			wp.ResponseBy = email
			wp.ResponseTime = &now

			// Update Order stage to "operations"
			if errStage := tx.Model(&entity.Order{}).Where("id = ?", wp.OrderID).Update("tahapan_proyek", "operations").Error; errStage != nil {
				return fmt.Errorf("failed to transition order stage to operations: %w", errStage)
			}
		} else {
			wp.Status = "draft"
		}

		wp.Stages = nil
		if errSaveWp := tx.Save(wp).Error; errSaveWp != nil {
			return fmt.Errorf("failed to save workplan record: %w", errSaveWp)
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	if req.Status == "submitted" {
		_ = s.logTaskSvc.TransitionStage(ctx, wp.OrderID, "operations", email)
	} else {
		_ = s.logTaskSvc.RecordTouch(ctx, wp.OrderID, "workplan", email)
	}

	return s.GetByID(ctx, wp.ID)
}

func (s *workplanService) toWorkplanResponse(wp entity.Workplan) *dto.WorkplanResponse {
	// Dedup stages by (InputItemRoomID, StageMasterID) to prevent layout duplication
	seen := make(map[string]bool)
	var uniqueStages []entity.WorkplanStage
	for _, stage := range wp.Stages {
		key := fmt.Sprintf("%d-%d", stage.InputItemRoomID, stage.StageMasterID)
		if !seen[key] {
			seen[key] = true
			uniqueStages = append(uniqueStages, stage)
		}
	}

	stages := make([]dto.WorkplanStageResponse, len(uniqueStages))
	for i, stage := range uniqueStages {
		var master *dto.WorkplanStageMasterResponse
		if stage.StageMaster != nil {
			master = &dto.WorkplanStageMasterResponse{
				ID:         stage.StageMaster.ID,
				Code:       stage.StageMaster.Code,
				Name:       stage.StageMaster.Name,
				Percentage: stage.StageMaster.Percentage,
				SortOrder:  stage.StageMaster.SortOrder,
				CreatedAt:  stage.StageMaster.CreatedAt,
				UpdatedAt:  stage.StageMaster.UpdatedAt,
			}
		}

		var roomName, prodName, prodDims string
		var qty int
		if stage.InputItemRoom != nil {
			roomName = stage.InputItemRoom.NamaRuangan
			qty = stage.InputItemRoom.Qty
			if stage.InputItemRoom.Produk != nil {
				prodName = stage.InputItemRoom.Produk.NamaProduk
			} else {
				prodName = "Kustom"
			}
			prodDims = fmt.Sprintf("%.1fx%.1fx%.1f", stage.InputItemRoom.Panjang, stage.InputItemRoom.Lebar, stage.InputItemRoom.Tinggi)
		}

		stages[i] = dto.WorkplanStageResponse{
			ID:              stage.ID,
			WorkplanID:      stage.WorkplanID,
			InputItemRoomID: stage.InputItemRoomID,
			StageMasterID:   stage.StageMasterID,
			Percentage:      stage.Percentage,
			StartDate:       stage.StartDate,
			EndDate:         stage.EndDate,
			Notes:           stage.Notes,
			Status:          stage.Status,
			CompletedAt:     stage.CompletedAt,
			CompletedBy:     stage.CompletedBy,
			Photos:          stage.Photos,
			CreatedAt:       stage.CreatedAt,
			UpdatedAt:       stage.UpdatedAt,
			RoomName:        roomName,
			ProductName:     prodName,
			ProductDims:     prodDims,
			Qty:             qty,
			StageMaster:     master,
		}
	}

	isFullyPaid := false
	var invoices []entity.Invoice
	if err := s.db.Where("order_id = ?", wp.OrderID).Find(&invoices).Error; err == nil {
		if len(invoices) > 0 {
			isFullyPaid = true
			for _, inv := range invoices {
				if strings.ToLower(inv.Keterangan) == "bast" {
					continue
				}
				if inv.Status != "terbayar" {
					isFullyPaid = false
					break
				}
			}
		}
	}

	return &dto.WorkplanResponse{
		ID:              wp.ID,
		OrderID:         wp.OrderID,
		Status:          wp.Status,
		StartDate:       wp.StartDate,
		EndDate:         wp.EndDate,
		DurationDays:    wp.DurationDays,
		ResponseBy:      wp.ResponseBy,
		ResponseTime:    wp.ResponseTime,
		ExtensionStatus: wp.ExtensionStatus,
		ExtensionNotes:  wp.ExtensionNotes,
		ExtensionDays:   wp.ExtensionDays,
		CreatedAt:       wp.CreatedAt,
		UpdatedAt:       wp.UpdatedAt,
		Stages:          stages,
		BastPhoto:       wp.BastPhoto,
		BastGeneratedAt: wp.BastGeneratedAt,
		BastGeneratedBy: wp.BastGeneratedBy,
		IsFullyPaid:     isFullyPaid,
	}
}

func (s *workplanService) ExportExcel(ctx context.Context, id uint) ([]byte, string, error) {
	wp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// Fetch associated Order contract date ranges
	var order entity.Order
	if errOrder := s.db.WithContext(ctx).First(&order, wp.OrderID).Error; errOrder != nil {
		return nil, "", errOrder
	}

	// Fetch input item rooms to get ordered list of products
	var inputItem entity.InputItem
	if errII := s.db.WithContext(ctx).Preload("Rooms.Produk").Where("order_id = ?", wp.OrderID).First(&inputItem).Error; errII != nil {
		return nil, "", errII
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Workplan"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Styles
	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "008080", Size: 14},
	})
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"008080"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleBody, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Size: 9},
		Border: []excelize.Border{
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	styleBodyC, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Size: 9},
		Border: []excelize.Border{
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleBold, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})

	cp := entity.GetCompanyProfile(s.db, order.CompanyID)
	_ = f.SetCellValue(sheetName, "A1", cp.Name)
	_ = f.SetCellStyle(sheetName, "A1", "A1", styleTitle)

	_ = f.SetCellValue(sheetName, "A2", "WORKPLAN PRODUKSI & TIMELINE")
	
	_ = f.SetCellValue(sheetName, "A4", "Nama Project:")
	_ = f.SetCellValue(sheetName, "B4", order.NamaProject)
	_ = f.SetCellValue(sheetName, "A5", "Nomor Order:")
	_ = f.SetCellValue(sheetName, "B5", order.NomorOrder)
	_ = f.SetCellValue(sheetName, "A6", "Nama Klien:")
	_ = f.SetCellValue(sheetName, "B6", order.NamaCustomer)

	_ = f.SetCellValue(sheetName, "D4", "Target Mulai:")
	if wp.StartDate != nil {
		_ = f.SetCellValue(sheetName, "E4", wp.StartDate.Format("2006-01-02"))
	} else {
		_ = f.SetCellValue(sheetName, "E4", "-")
	}
	_ = f.SetCellValue(sheetName, "D5", "Target Selesai:")
	if wp.EndDate != nil {
		_ = f.SetCellValue(sheetName, "E5", wp.EndDate.Format("2006-01-02"))
	} else {
		_ = f.SetCellValue(sheetName, "E5", "-")
	}
	_ = f.SetCellValue(sheetName, "D6", "Durasi Proyek:")
	_ = f.SetCellValue(sheetName, "E6", fmt.Sprintf("%d Hari", wp.DurationDays))

	_ = f.SetCellStyle(sheetName, "A4", "A6", styleBold)
	_ = f.SetCellStyle(sheetName, "D4", "D6", styleBold)

	headers := []string{"NO.", "RUANGAN", "PRODUK", "TAHAPAN", "PERSENTASE (%)", "TANGGAL MULAI", "TANGGAL SELESAI", "DURASI (HARI)", "CATATAN / NOTES"}
	startRow := 8
	colLetter := func(col int) string {
		return string(rune('A' + col - 1))
	}

	for colIdx, header := range headers {
		cell := fmt.Sprintf("%s%d", colLetter(colIdx+1), startRow)
		_ = f.SetCellValue(sheetName, cell, header)
		_ = f.SetCellStyle(sheetName, cell, cell, styleHeader)
	}

	// Map stage master ID to stages for quick lookup (deduplicated)
	productStagesMap := make(map[uint][]entity.WorkplanStage)
	seen := make(map[string]bool)
	for _, stage := range wp.Stages {
		key := fmt.Sprintf("%d-%d", stage.InputItemRoomID, stage.StageMasterID)
		if !seen[key] {
			seen[key] = true
			productStagesMap[stage.InputItemRoomID] = append(productStagesMap[stage.InputItemRoomID], stage)
		}
	}

	// Sort stages inside each product by sort_order of master stage
	for roomId := range productStagesMap {
		stagesList := productStagesMap[roomId]
		// Sort by sort_order
		for i := 0; i < len(stagesList); i++ {
			for j := i + 1; j < len(stagesList); j++ {
				iOrder, jOrder := 0, 0
				if stagesList[i].StageMaster != nil {
					iOrder = stagesList[i].StageMaster.SortOrder
				}
				if stagesList[j].StageMaster != nil {
					jOrder = stagesList[j].StageMaster.SortOrder
				}
				if iOrder > jOrder {
					stagesList[i], stagesList[j] = stagesList[j], stagesList[i]
				}
			}
		}
		productStagesMap[roomId] = stagesList
	}

	curRow := startRow + 1
	noCount := 1

	for _, room := range inputItem.Rooms {
		if room.ProdukID == nil {
			continue
		}
		stages := productStagesMap[room.ID]
		if len(stages) == 0 {
			continue
		}

		prodStartRow := curRow
		prodName := room.Produk.NamaProduk
		if prodName == "" {
			prodName = "Kustom"
		}
		prodSpec := fmt.Sprintf("%s (%.1fx%.1fx%.1f)", prodName, room.Panjang, room.Lebar, room.Tinggi)

		for _, stage := range stages {
			// Col 4: Tahapan
			stageName := ""
			if stage.StageMaster != nil {
				stageName = stage.StageMaster.Name
			}
			_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", curRow), stageName)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("D%d", curRow), fmt.Sprintf("D%d", curRow), styleBody)

			// Col 5: Persentase
			_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", curRow), fmt.Sprintf("%.1f%%", stage.Percentage))
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("E%d", curRow), fmt.Sprintf("E%d", curRow), styleBodyC)

			// Col 6: Start Date
			if stage.StartDate != nil {
				_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", curRow), stage.StartDate.Format("2006-01-02"))
			} else {
				_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", curRow), "-")
			}
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("F%d", curRow), fmt.Sprintf("F%d", curRow), styleBodyC)

			// Col 7: End Date
			if stage.EndDate != nil {
				_ = f.SetCellValue(sheetName, fmt.Sprintf("G%d", curRow), stage.EndDate.Format("2006-01-02"))
			} else {
				_ = f.SetCellValue(sheetName, fmt.Sprintf("G%d", curRow), "-")
			}
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("G%d", curRow), fmt.Sprintf("G%d", curRow), styleBodyC)

			// Col 8: Duration
			duration := "-"
			if stage.StartDate != nil && stage.EndDate != nil {
				days := int(stage.EndDate.Sub(*stage.StartDate).Hours()/24) + 1
				duration = fmt.Sprintf("%d", days)
			}
			_ = f.SetCellValue(sheetName, fmt.Sprintf("H%d", curRow), duration)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("H%d", curRow), fmt.Sprintf("H%d", curRow), styleBodyC)

			// Col 9: Notes
			_ = f.SetCellValue(sheetName, fmt.Sprintf("I%d", curRow), stage.Notes)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("I%d", curRow), fmt.Sprintf("I%d", curRow), styleBody)

			// Add cell borders to A, B, C rows (which will be merged later)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("A%d", curRow), styleBodyC)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("B%d", curRow), fmt.Sprintf("B%d", curRow), styleBody)
			_ = f.SetCellStyle(sheetName, fmt.Sprintf("C%d", curRow), fmt.Sprintf("C%d", curRow), styleBody)

			curRow++
		}

		// Merge product row metadata
		if curRow > prodStartRow {
			_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", prodStartRow), fmt.Sprintf("A%d", curRow-1))
			_ = f.MergeCell(sheetName, fmt.Sprintf("B%d", prodStartRow), fmt.Sprintf("B%d", curRow-1))
			_ = f.MergeCell(sheetName, fmt.Sprintf("C%d", prodStartRow), fmt.Sprintf("C%d", curRow-1))

			_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", prodStartRow), noCount)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", prodStartRow), room.NamaRuangan)
			_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", prodStartRow), prodSpec)

			noCount++
		}
	}

	// Widths
	widths := map[string]float64{
		"A": 6,
		"B": 24,
		"C": 36,
		"D": 18,
		"E": 14,
		"F": 18,
		"G": 18,
		"H": 14,
		"I": 30,
	}
	for col, w := range widths {
		_ = f.SetColWidth(sheetName, col, col, w)
	}

	buf, errXls := f.WriteToBuffer()
	if errXls != nil {
		return nil, "", errXls
	}

	filename := fmt.Sprintf("Workplan_%s.xlsx", order.NomorOrder)
	return buf.Bytes(), filename, nil
}

func (s *workplanService) GetStageMasters(ctx context.Context) ([]dto.WorkplanStageMasterResponse, error) {
	masters, err := s.repo.GetStageMasters(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.WorkplanStageMasterResponse, len(masters))
	for i, m := range masters {
		result[i] = dto.WorkplanStageMasterResponse{
			ID:         m.ID,
			Code:       m.Code,
			Name:       m.Name,
			Percentage: m.Percentage,
			SortOrder:  m.SortOrder,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		}
	}
	return result, nil
}

func (s *workplanService) UpdateStageMasters(ctx context.Context, req []dto.UpdateWorkplanStageMasterRequest) error {
	var totalPercentage float64
	for _, item := range req {
		totalPercentage += item.Percentage
	}
	// Round to 2 decimal places to avoid precision issues
	totalPercentage = float64(int(totalPercentage*100+0.5)) / 100
	if totalPercentage != 100.00 {
		return fmt.Errorf("total persentase tahapan harus tepat 100%% (saat ini %.2f%%)", totalPercentage)
	}

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reqIds []uint
		for _, item := range req {
			if item.ID > 0 {
				reqIds = append(reqIds, item.ID)
			}
		}

		if len(reqIds) > 0 {
			if err := tx.Where("id NOT IN ?", reqIds).Delete(&entity.WorkplanStageMaster{}).Error; err != nil {
				return fmt.Errorf("gagal menghapus templat tahapan lama: %w", err)
			}
		} else {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.WorkplanStageMaster{}).Error; err != nil {
				return fmt.Errorf("gagal mengosongkan templat tahapan: %w", err)
			}
		}

		for i, item := range req {
			master := entity.WorkplanStageMaster{
				Code:       item.Code,
				Name:       item.Name,
				Percentage: item.Percentage,
				SortOrder:  i + 1,
			}
			if item.ID > 0 {
				master.ID = item.ID
				if err := tx.Save(&master).Error; err != nil {
					return fmt.Errorf("gagal menyimpan templat tahapan '%s': %w", item.Name, err)
				}
			} else {
				if err := tx.Create(&master).Error; err != nil {
					return fmt.Errorf("gagal membuat templat tahapan '%s': %w", item.Name, err)
				}
			}
		}
		return nil
	})

	return txErr
}

func (s *workplanService) CompleteStage(ctx context.Context, stageID uint, req dto.CompleteStageRequest, userEmail string) error {
	if len(req.Photos) == 0 {
		return errors.New("wajib mengunggah minimal satu foto bukti")
	}

	// 1. Get the target stage
	var targetStage entity.WorkplanStage
	err := s.db.WithContext(ctx).
		Preload("StageMaster").
		First(&targetStage, stageID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("tahapan tidak ditemukan")
		}
		return err
	}

	if targetStage.Status == "completed" {
		return errors.New("tahapan ini sudah diselesaikan")
	}

	// Check active defect
	var activeDefect entity.WorkplanDefect
	if err := s.db.WithContext(ctx).Where("workplan_stage_id = ? AND status IN ?", stageID, []string{"reported", "fix_submitted"}).First(&activeDefect).Error; err == nil {
		return errors.New("tahapan ini memiliki laporan defect yang belum diselesaikan. Selesaikan defect terlebih dahulu")
	}

	// 2. Load all sibling stages for the same WorkplanID and InputItemRoomID in order
	var orderedStages []entity.WorkplanStage
	err = s.db.WithContext(ctx).
		Joins("JOIN workplan_stage_masters ON workplan_stage_masters.id = workplan_stages.stage_master_id").
		Preload("StageMaster").
		Where("workplan_stages.workplan_id = ? AND workplan_stages.input_item_room_id = ?", targetStage.WorkplanID, targetStage.InputItemRoomID).
		Order("workplan_stage_masters.sort_order ASC").
		Find(&orderedStages).Error
	if err != nil {
		return err
	}

	// 3. Check sequence
	foundTarget := false
	for _, stage := range orderedStages {
		if stage.ID == targetStage.ID {
			foundTarget = true
			break
		}
		if stage.Status != "completed" {
			return fmt.Errorf("tahapan sebelumnya '%s' belum diselesaikan", stage.StageMaster.Name)
		}
	}

	if !foundTarget {
		return errors.New("tahapan tidak valid dalam urutan produk")
	}

	// 4. Update stage
	now := time.Now()
	targetStage.Status = "completed"
	targetStage.CompletedAt = &now
	targetStage.CompletedBy = userEmail
	targetStage.Notes = req.Notes
	
	importStrings := ""
	for i, ph := range req.Photos {
		if i > 0 {
			importStrings += ","
		}
		importStrings += ph
	}
	targetStage.Photos = importStrings

	err = s.db.WithContext(ctx).Save(&targetStage).Error
	if err != nil {
		return fmt.Errorf("gagal menyimpan status tahapan: %w", err)
	}

	return nil
}

func (s *workplanService) RequestExtension(ctx context.Context, wpID uint, req dto.RequestExtensionRequest) (*dto.WorkplanResponse, error) {
	wp, err := s.repo.FindByID(ctx, wpID)
	if err != nil {
		return nil, err
	}
	if wp.Status != "submitted" {
		return nil, errors.New("hanya workplan yang telah disubmit yang dapat diajukan perpanjangan timeline")
	}

	wp.ExtensionStatus = "pending"
	wp.ExtensionNotes = req.Notes
	wp.ExtensionDays = req.Days

	if err := s.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, wp.ID)
}

func (s *workplanService) HandleExtension(ctx context.Context, wpID uint, req dto.HandleExtensionRequest) (*dto.WorkplanResponse, error) {
	wp, err := s.repo.FindByID(ctx, wpID)
	if err != nil {
		return nil, err
	}
	if wp.ExtensionStatus != "pending" {
		return nil, errors.New("tidak ada pengajuan perpanjangan timeline yang pending")
	}

	if req.Action == "approve" {
		wp.ExtensionStatus = "approved"
		wp.Status = "draft" // unlock!

		if wp.EndDate != nil && wp.ExtensionDays > 0 {
			newEnd := wp.EndDate.AddDate(0, 0, wp.ExtensionDays)
			wp.EndDate = &newEnd
			wp.DurationDays = wp.DurationDays + wp.ExtensionDays
		}
	} else {
		wp.ExtensionStatus = "rejected"
	}

	if err := s.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, wp.ID)
}

func (s *workplanService) resolvePhotoPath(photoURL string) string {
	cleanPath := strings.TrimPrefix(photoURL, "/uploads")
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	return filepath.Join(s.uploadDir, cleanPath)
}

func drawCheckmark(pdf *gofpdf.Fpdf, x, y, w, h float64) {
	pdf.SetDrawColor(46, 117, 89) // dark green
	pdf.SetLineWidth(0.4)
	cx := x + w/2
	cy := y + h/2
	pdf.Line(cx-2.2, cy, cx-0.5, cy+1.8)
	pdf.Line(cx-0.5, cy+1.8, cx+2.2, cy-1.8)
}

func truncateString(str string, maxLen int) string {
	runes := []rune(str)
	if len(runes) > maxLen {
		if maxLen > 3 {
			return string(runes[:maxLen-3]) + "..."
		}
		return string(runes[:maxLen])
	}
	return str
}

func calculateProjectProgress(stages []entity.WorkplanStage) float64 {
	if len(stages) == 0 {
		return 0
	}
	productStages := make(map[uint][]entity.WorkplanStage)
	for _, stg := range stages {
		productStages[stg.InputItemRoomID] = append(productStages[stg.InputItemRoomID], stg)
	}
	if len(productStages) == 0 {
		return 0
	}
	var totalProductProgress float64
	for _, stgs := range productStages {
		var sumAllWeights float64
		var sumCompletedWeights float64
		for _, s := range stgs {
			sumAllWeights += s.Percentage
			if s.Status == "completed" {
				sumCompletedWeights += s.Percentage
			}
		}
		productProgress := 0.0
		if sumAllWeights > 0 {
			productProgress = (sumCompletedWeights / sumAllWeights) * 100
		}
		totalProductProgress += productProgress
	}
	return totalProductProgress / float64(len(productStages))
}

func (s *workplanService) ExportProgressPDF(ctx context.Context, id uint) ([]byte, string, error) {
	wp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	var order entity.Order
	if errOrder := s.db.WithContext(ctx).Preload("PIC").Preload("Contracts").First(&order, wp.OrderID).Error; errOrder != nil {
		return nil, "", errOrder
	}

	var inputItem entity.InputItem
	if errII := s.db.WithContext(ctx).Preload("Rooms.Produk").Where("order_id = ?", wp.OrderID).First(&inputItem).Error; errII != nil {
		return nil, "", errII
	}

	var stageMasters []entity.WorkplanStageMaster
	if errMasters := s.db.WithContext(ctx).Order("sort_order ASC").Find(&stageMasters).Error; errMasters != nil {
		return nil, "", errMasters
	}

	// Calculate target metadata
	var tanggalKontrakStr string = "-"
	var approvalGambarKerjaStr string = "-"
	var jangkaWaktuStr string = "-"
	var target50Str string = "-"
	var target80Str string = "-"
	var target100Str string = "-"

	var startDate time.Time
	hasStartDate := false

	if order.TanggalKontrak != nil {
		startDate = *order.TanggalKontrak
		hasStartDate = true
		tanggalKontrakStr = order.TanggalKontrak.Format("02/01/2006")
	} else if wp.StartDate != nil {
		startDate = *wp.StartDate
		hasStartDate = true
		tanggalKontrakStr = wp.StartDate.Format("02/01/2006")
	}

	var gk entity.GambarKerja
	if errGK := s.db.WithContext(ctx).Where("order_id = ? AND status = 'approved'", wp.OrderID).First(&gk).Error; errGK == nil && gk.ResponseTime != nil {
		approvalGambarKerjaStr = gk.ResponseTime.Format("02/01/2006")
	} else if order.TanggalKontrak != nil {
		approvalGambarKerjaStr = order.TanggalKontrak.Format("02/01/2006")
	} else if wp.ResponseTime != nil {
		approvalGambarKerjaStr = wp.ResponseTime.Format("02/01/2006")
	}

	durationDays := wp.DurationDays
	if len(order.Contracts) > 0 {
		jangkaWaktuStr = order.Contracts[0].LamaKontrak
		if order.TanggalMulai != nil && order.TanggalSelesai != nil {
			days := int(order.TanggalSelesai.Sub(*order.TanggalMulai).Hours()/24) + 1
			durationDays = days
			jangkaWaktuStr = fmt.Sprintf("%d Hari", days)
		}
	} else {
		jangkaWaktuStr = fmt.Sprintf("%d Hari", durationDays)
	}

	if durationDays <= 0 {
		durationDays = 60
	}

	if hasStartDate {
		days50 := int(float64(durationDays) * 0.5)
		days80 := int(float64(durationDays) * 0.8)
		days100 := durationDays

		t50 := startDate.AddDate(0, 0, days50)
		t80 := startDate.AddDate(0, 0, days80)
		t100 := startDate.AddDate(0, 0, days100)

		target50Str = t50.Format("02/01/2006")
		target80Str = t80.Format("02/01/2006")
		target100Str = t100.Format("02/01/2006")
	}

	// Fetch RAB to determine bobot per product
	var rab entity.RAB
	hasRAB := false
	if errRAB := s.db.WithContext(ctx).Preload("Rooms").Where("order_id = ?", wp.OrderID).First(&rab).Error; errRAB == nil {
		hasRAB = true
	}

	rabRoomPrices := make(map[string]float64)
	if hasRAB {
		for _, rr := range rab.Rooms {
			prodID := uint(0)
			if rr.ProdukID != nil {
				prodID = *rr.ProdukID
			}
			key := fmt.Sprintf("%s-%d-%.1f-%.1f-%.1f", rr.NamaRuangan, prodID, rr.Panjang, rr.Lebar, rr.Tinggi)
			rabRoomPrices[key] = rr.HargaTotal
		}
	}

	// Map stage master ID to stages for quick lookup
	productStagesMap := make(map[uint]map[uint]entity.WorkplanStage)
	for _, stage := range wp.Stages {
		if productStagesMap[stage.InputItemRoomID] == nil {
			productStagesMap[stage.InputItemRoomID] = make(map[uint]entity.WorkplanStage)
		}
		productStagesMap[stage.InputItemRoomID][stage.StageMasterID] = stage
	}

	// Landscape A4 PDF
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title / Header
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 102, 102) // Dark Teal
	pdf.CellFormat(0, 6, "PROJECT PROGRESS REPORT", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 4, fmt.Sprintf("%s — %s", order.NamaProject, order.NamaCustomer), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// Draw metadata box
	pdf.SetFont("Arial", "B", 7.5)
	pdf.SetDrawColor(220, 225, 225)
	pdf.SetLineWidth(0.2)

	// Row 1
	pdf.SetFillColor(245, 247, 248)
	pdf.CellFormat(40, 6, "Tanggal Kontrak", "1", 0, "L", true, 0, "")
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(45, 6, tanggalKontrakStr, "1", 0, "L", false, 0, "")

	pdf.SetFillColor(245, 247, 248)
	pdf.CellFormat(25, 6, "Customer", "1", 0, "L", true, 0, "")

	xCustVal := pdf.GetX()
	pdf.SetFillColor(255, 242, 242)
	pdf.SetTextColor(200, 30, 30)
	pdf.SetFont("Arial", "B", 10.5)
	pdf.CellFormat(72, 18, order.NamaCustomer, "1", 0, "C", true, 0, "")

	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Arial", "B", 7.5)
	pdf.SetFillColor(245, 247, 248)
	pdf.CellFormat(45, 6, "Target Progres 50%", "1", 0, "L", true, 0, "")
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(40, 6, target50Str, "1", 1, "L", false, 0, "")

	// Row 2
	pdf.SetFillColor(245, 247, 248)
	pdf.CellFormat(40, 6, "Approval Gambar Kerja", "1", 0, "L", true, 0, "")
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(45, 6, approvalGambarKerjaStr, "1", 0, "L", false, 0, "")

	pdf.SetX(xCustVal + 72)

	pdf.SetFillColor(245, 247, 248)
	pdf.CellFormat(45, 6, "Target Progres 80%", "1", 0, "L", true, 0, "")
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(40, 6, target80Str, "1", 1, "L", false, 0, "")

	// Row 3
	pdf.SetFillColor(245, 247, 248)
	pdf.CellFormat(40, 6, "Jangka Waktu", "1", 0, "L", true, 0, "")
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(45, 6, jangkaWaktuStr, "1", 0, "L", false, 0, "")

	pdf.SetX(xCustVal + 72)

	pdf.SetFillColor(245, 247, 248)
	pdf.CellFormat(45, 6, "Target Progres 100%", "1", 0, "L", true, 0, "")
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(40, 6, target100Str, "1", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Draw Total Progress Bar
	yBar := pdf.GetY()
	pdf.SetFillColor(242, 247, 247)
	pdf.Rect(15, yBar, 267, 7, "F")

	pdf.SetTextColor(0, 102, 102)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetXY(18, yBar+1.5)
	pdf.CellFormat(25, 4, "Total Progress:", "", 0, "L", false, 0, "")

	xProgressStart := 43.0
	progressBarWidth := 90.0
	pdf.SetFillColor(220, 225, 225)
	pdf.Rect(xProgressStart, yBar+2, progressBarWidth, 3, "F")

	totalProgress := calculateProjectProgress(wp.Stages)
	if totalProgress > 0 {
		pdf.SetFillColor(0, 128, 128)
		pdf.Rect(xProgressStart, yBar+2, progressBarWidth*(totalProgress/100.0), 3, "F")
	}

	pdf.SetXY(xProgressStart+progressBarWidth+4, yBar+1.5)
	pdf.CellFormat(20, 4, fmt.Sprintf("%.2f%%", totalProgress), "", 0, "L", false, 0, "")

	todayStr := time.Now().Format("02/01/2006")
	pdf.SetXY(230, yBar+1.5)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(0, 4, "Hari ini: "+todayStr, "", 1, "R", false, 0, "")
	pdf.Ln(4)

	// Table Header
	pdf.SetFillColor(20, 30, 50)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 7)
	pdf.SetDrawColor(200, 200, 200)

	xStartTable := pdf.GetX()
	yStartTable := pdf.GetY()

	pdf.CellFormat(7, 10, "NO", "1", 0, "C", true, 0, "")
	pdf.CellFormat(36, 10, "JENIS PEKERJAAN", "1", 0, "C", true, 0, "")
	pdf.CellFormat(10, 10, "QTY", "1", 0, "C", true, 0, "")
	pdf.CellFormat(12, 10, "SATUAN", "1", 0, "C", true, 0, "")
	pdf.CellFormat(62, 10, "MATERIAL", "1", 0, "C", true, 0, "")
	pdf.CellFormat(15, 10, "BOBOT", "1", 0, "C", true, 0, "")

	numStages := len(stageMasters)
	if numStages == 0 {
		numStages = 1
	}
	stageColWidth := 125.0 / float64(numStages)

	xStageStart := pdf.GetX()
	pdf.CellFormat(125, 5, "TAHAPAN PRODUKSI", "1", 1, "C", true, 0, "")

	pdf.SetXY(xStageStart, yStartTable+5)
	pdf.SetFont("Arial", "B", 5.2)
	for _, m := range stageMasters {
		xCell := pdf.GetX()
		yCell := pdf.GetY()
		pdf.CellFormat(stageColWidth, 2.5, truncateString(m.Name, 7), "1", 0, "C", true, 0, "")
		pdf.SetXY(xCell, yCell+2.5)
		pdf.CellFormat(stageColWidth, 2.5, fmt.Sprintf("%.0f%%", m.Percentage), "1", 0, "C", true, 0, "")
		pdf.SetXY(xCell+stageColWidth, yCell)
	}
	pdf.SetXY(xStartTable, yStartTable+10)

	// Group rooms
	groupedRooms := make(map[string][]entity.InputItemRoom)
	var roomOrder []string
	seenRooms := make(map[string]bool)

	for _, r := range inputItem.Rooms {
		if !seenRooms[r.NamaRuangan] {
			seenRooms[r.NamaRuangan] = true
			roomOrder = append(roomOrder, r.NamaRuangan)
		}
		groupedRooms[r.NamaRuangan] = append(groupedRooms[r.NamaRuangan], r)
	}

	noCount := 1
	for _, rName := range roomOrder {
		// Draw Room separator bar
		pdf.SetFillColor(52, 144, 242) // Sky blue background
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(267, 6, strings.ToUpper(rName), "1", 1, "L", true, 0, "")

		pdf.SetTextColor(50, 50, 50)
		pdf.SetFont("Arial", "", 7)
		pdf.SetDrawColor(230, 235, 235)

		for rowIdx, room := range groupedRooms[rName] {
			pdf.SetFillColor(255, 255, 255)
			if rowIdx%2 == 1 {
				pdf.SetFillColor(248, 250, 252)
			}

			prodID := uint(0)
			if room.ProdukID != nil {
				prodID = *room.ProdukID
			}
			key := fmt.Sprintf("%s-%d-%.1f-%.1f-%.1f", room.NamaRuangan, prodID, room.Panjang, room.Lebar, room.Tinggi)
			hargaTotal := rabRoomPrices[key]

			bobot := 0.0
			if hasRAB && rab.GrandTotal > 0 {
				bobot = (hargaTotal / rab.GrandTotal) * 100
			} else {
				bobot = 100.0 / float64(len(inputItem.Rooms))
			}

			prodName := "Kustom"
			if room.Produk != nil {
				prodName = room.Produk.NamaProduk
			}

			materialStr := ""
			for mIdx, bb := range room.BahanBakus {
				if bb.BahanBaku != nil {
					if mIdx > 0 {
						materialStr += ", "
					}
					materialStr += bb.BahanBaku.NamaBahanBaku
				}
			}

			pdf.CellFormat(7, 6, fmt.Sprintf("%d", noCount), "1", 0, "C", true, 0, "")
			pdf.CellFormat(36, 6, truncateString(prodName, 22), "1", 0, "L", true, 0, "")
			pdf.CellFormat(10, 6, fmt.Sprintf("%d", room.Qty), "1", 0, "C", true, 0, "")
			pdf.CellFormat(12, 6, "Unit", "1", 0, "L", true, 0, "")
			pdf.CellFormat(62, 6, truncateString(materialStr, 40), "1", 0, "L", true, 0, "")
			pdf.CellFormat(15, 6, fmt.Sprintf("%.2f%%", bobot), "1", 0, "R", true, 0, "")

			for _, m := range stageMasters {
				xCell := pdf.GetX()
				yCell := pdf.GetY()

				stage, stageExists := productStagesMap[room.ID][m.ID]
				pdf.CellFormat(stageColWidth, 6, "", "1", 0, "C", true, 0, "")

				if stageExists && stage.Status == "completed" {
					drawCheckmark(pdf, xCell, yCell, stageColWidth, 6)
				}
			}
			pdf.Ln(6)
			noCount++
		}
	}

	// Halaman lampiran foto dokumentasi jika ada foto
	hasPhotos := false
	for _, rName := range roomOrder {
		for _, room := range groupedRooms[rName] {
			for _, m := range stageMasters {
				if stg, exists := productStagesMap[room.ID][m.ID]; exists && stg.Photos != "" {
					hasPhotos = true
					break
				}
			}
		}
	}

	if hasPhotos {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.SetTextColor(0, 102, 102)
		pdf.CellFormat(0, 8, "DOKUMENTASI TAHAPAN PRODUKSI", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8.5)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 4, fmt.Sprintf("%s — %s", order.NamaProject, order.NamaCustomer), "", 1, "L", false, 0, "")
		pdf.Ln(6)

		for _, rName := range roomOrder {
			roomHasPhotos := false
			for _, room := range groupedRooms[rName] {
				for _, m := range stageMasters {
					if stg, exists := productStagesMap[room.ID][m.ID]; exists && stg.Photos != "" {
						roomHasPhotos = true
						break
					}
				}
			}
			if !roomHasPhotos {
				continue
			}

			pdf.SetFillColor(52, 144, 242)
			pdf.SetTextColor(255, 255, 255)
			pdf.SetFont("Arial", "B", 9)
			pdf.CellFormat(267, 7, "RUANGAN: "+strings.ToUpper(rName), "1", 1, "L", true, 0, "")
			pdf.Ln(4)

			for _, room := range groupedRooms[rName] {
				prodHasPhotos := false
				for _, m := range stageMasters {
					if stg, exists := productStagesMap[room.ID][m.ID]; exists && stg.Photos != "" {
						prodHasPhotos = true
						break
					}
				}
				if !prodHasPhotos {
					continue
				}

				prodName := "Kustom"
				if room.Produk != nil {
					prodName = room.Produk.NamaProduk
				}

				pdf.SetTextColor(0, 102, 102)
				pdf.SetFont("Arial", "B", 8.5)
				pdf.CellFormat(267, 5, fmt.Sprintf("%s (Qty: %d)", prodName, room.Qty), "B", 1, "L", false, 0, "")
				pdf.Ln(3)

				xStart := 15.0
				cardW := 60.0
				cardH := 50.0
				spacing := 9.0

				colCount := 0
				yStartRow := pdf.GetY()

				for _, m := range stageMasters {
					stage, exists := productStagesMap[room.ID][m.ID]
					if !exists || stage.Photos == "" {
						continue
					}
					photosList := strings.Split(stage.Photos, ",")
					if len(photosList) == 0 || photosList[0] == "" {
						continue
					}

					if yStartRow+cardH > 195 {
						pdf.AddPage()
						yStartRow = 15.0
					}

					xVal := xStart + float64(colCount)*(cardW+spacing)
					yVal := yStartRow

					pdf.SetDrawColor(220, 225, 225)
					pdf.SetLineWidth(0.2)
					pdf.Rect(xVal, yVal, cardW, cardH, "D")

					fullImgPath := s.resolvePhotoPath(photosList[0])
					if _, errStat := os.Stat(fullImgPath); errStat == nil {
						pdf.Image(fullImgPath, xVal+1, yVal+1, cardW-2, 30, false, "", 0, "")
					} else {
						pdf.SetFillColor(240, 242, 245)
						pdf.Rect(xVal+1, yVal+1, cardW-2, 30, "F")
						pdf.SetTextColor(150, 150, 150)
						pdf.SetFont("Arial", "I", 8)
						pdf.SetXY(xVal+1, yVal+13)
						pdf.CellFormat(cardW-2, 5, "Foto tidak ditemukan", "", 0, "C", false, 0, "")
					}

					pdf.SetTextColor(50, 50, 50)
					pdf.SetFont("Arial", "B", 7.5)
					pdf.SetXY(xVal+1, yVal+32)
					pdf.CellFormat(cardW-2, 3.5, stage.StageMaster.Name, "", 1, "C", false, 0, "")

					pdf.SetFont("Arial", "", 6.5)
					pdf.SetXY(xVal+1, yVal+35.5)
					pdf.CellFormat(cardW-2, 3, "Oleh: "+stage.CompletedBy, "", 1, "C", false, 0, "")

					timeStr := ""
					if stage.CompletedAt != nil {
						timeStr = stage.CompletedAt.Format("02/01/2006 15:04")
					}
					pdf.SetXY(xVal+1, yVal+38.5)
					pdf.CellFormat(cardW-2, 3, timeStr, "", 1, "C", false, 0, "")

					pdf.SetFont("Arial", "I", 6.5)
					pdf.SetTextColor(100, 100, 100)
					pdf.SetXY(xVal+1, yVal+42)
					pdf.CellFormat(cardW-2, 3, truncateString(stage.Notes, 24), "", 1, "C", false, 0, "")

					colCount++
					if colCount >= 4 {
						colCount = 0
						yStartRow += cardH + spacing
						pdf.SetY(yStartRow)
					}
				}

				if colCount > 0 {
					yStartRow += cardH + spacing
					pdf.SetY(yStartRow)
				}
				pdf.Ln(4)
			}
		}
	}

	var buf bytes.Buffer
	if errPdf := pdf.Output(&buf); errPdf != nil {
		return nil, "", errPdf
	}

	filename := fmt.Sprintf("Progress_Report_%s.pdf", order.NomorOrder)
	return buf.Bytes(), filename, nil
}

func (s *workplanService) ExportProgressExcel(ctx context.Context, id uint) ([]byte, string, error) {
	wp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	var order entity.Order
	if errOrder := s.db.WithContext(ctx).Preload("PIC").Preload("Contracts").First(&order, wp.OrderID).Error; errOrder != nil {
		return nil, "", errOrder
	}

	var inputItem entity.InputItem
	if errII := s.db.WithContext(ctx).Preload("Rooms.Produk").Where("order_id = ?", wp.OrderID).First(&inputItem).Error; errII != nil {
		return nil, "", errII
	}

	var stageMasters []entity.WorkplanStageMaster
	if errMasters := s.db.WithContext(ctx).Order("sort_order ASC").Find(&stageMasters).Error; errMasters != nil {
		return nil, "", errMasters
	}

	// Calculate target metadata
	var tanggalKontrakStr string = "-"
	var approvalGambarKerjaStr string = "-"
	var jangkaWaktuStr string = "-"
	var target50Str string = "-"
	var target80Str string = "-"
	var target100Str string = "-"

	var startDate time.Time
	hasStartDate := false

	if order.TanggalKontrak != nil {
		startDate = *order.TanggalKontrak
		hasStartDate = true
		tanggalKontrakStr = order.TanggalKontrak.Format("02/01/2006")
	} else if wp.StartDate != nil {
		startDate = *wp.StartDate
		hasStartDate = true
		tanggalKontrakStr = wp.StartDate.Format("02/01/2006")
	}

	var gk entity.GambarKerja
	if errGK := s.db.WithContext(ctx).Where("order_id = ? AND status = 'approved'", wp.OrderID).First(&gk).Error; errGK == nil && gk.ResponseTime != nil {
		approvalGambarKerjaStr = gk.ResponseTime.Format("02/01/2006")
	} else if order.TanggalKontrak != nil {
		approvalGambarKerjaStr = order.TanggalKontrak.Format("02/01/2006")
	} else if wp.ResponseTime != nil {
		approvalGambarKerjaStr = wp.ResponseTime.Format("02/01/2006")
	}

	durationDays := wp.DurationDays
	if len(order.Contracts) > 0 {
		jangkaWaktuStr = order.Contracts[0].LamaKontrak
		if order.TanggalMulai != nil && order.TanggalSelesai != nil {
			days := int(order.TanggalSelesai.Sub(*order.TanggalMulai).Hours()/24) + 1
			durationDays = days
			jangkaWaktuStr = fmt.Sprintf("%d Hari", days)
		}
	} else {
		jangkaWaktuStr = fmt.Sprintf("%d Hari", durationDays)
	}

	if durationDays <= 0 {
		durationDays = 60
	}

	if hasStartDate {
		days50 := int(float64(durationDays) * 0.5)
		days80 := int(float64(durationDays) * 0.8)
		days100 := durationDays

		t50 := startDate.AddDate(0, 0, days50)
		t80 := startDate.AddDate(0, 0, days80)
		t100 := startDate.AddDate(0, 0, days100)

		target50Str = t50.Format("02/01/2006")
		target80Str = t80.Format("02/01/2006")
		target100Str = t100.Format("02/01/2006")
	}

	// Fetch RAB to determine bobot per product
	var rab entity.RAB
	hasRAB := false
	if errRAB := s.db.WithContext(ctx).Preload("Rooms").Where("order_id = ?", wp.OrderID).First(&rab).Error; errRAB == nil {
		hasRAB = true
	}

	rabRoomPrices := make(map[string]float64)
	if hasRAB {
		for _, rr := range rab.Rooms {
			prodID := uint(0)
			if rr.ProdukID != nil {
				prodID = *rr.ProdukID
			}
			key := fmt.Sprintf("%s-%d-%.1f-%.1f-%.1f", rr.NamaRuangan, prodID, rr.Panjang, rr.Lebar, rr.Tinggi)
			rabRoomPrices[key] = rr.HargaTotal
		}
	}

	// Map stage master ID to stages for quick lookup
	productStagesMap := make(map[uint]map[uint]entity.WorkplanStage)
	for _, stage := range wp.Stages {
		if productStagesMap[stage.InputItemRoomID] == nil {
			productStagesMap[stage.InputItemRoomID] = make(map[uint]entity.WorkplanStage)
		}
		productStagesMap[stage.InputItemRoomID][stage.StageMasterID] = stage
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet1 := "Progress Report"
	_ = f.SetSheetName("Sheet1", sheet1)

	// Styles
	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "008080", Size: 14},
	})
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 9},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"141E32"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleRoomRow, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"3490F2"}, Pattern: 1},
	})
	styleBody, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Size: 9},
		Border: []excelize.Border{
			{Type: "left", Color: "D0D5D5", Style: 1},
			{Type: "right", Color: "D0D5D5", Style: 1},
			{Type: "top", Color: "D0D5D5", Style: 1},
			{Type: "bottom", Color: "D0D5D5", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	styleBodyC, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Size: 9},
		Border: []excelize.Border{
			{Type: "left", Color: "D0D5D5", Style: 1},
			{Type: "right", Color: "D0D5D5", Style: 1},
			{Type: "top", Color: "D0D5D5", Style: 1},
			{Type: "bottom", Color: "D0D5D5", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	stylePinkBox, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "C81E1E", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FFF0F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "D0D5D5", Style: 1},
			{Type: "right", Color: "D0D5D5", Style: 1},
			{Type: "top", Color: "D0D5D5", Style: 1},
			{Type: "bottom", Color: "D0D5D5", Style: 1},
		},
	})
	styleLabel, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 9},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F5F7F8"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "D0D5D5", Style: 1},
			{Type: "right", Color: "D0D5D5", Style: 1},
			{Type: "top", Color: "D0D5D5", Style: 1},
			{Type: "bottom", Color: "D0D5D5", Style: 1},
		},
	})

	_ = f.SetCellValue(sheet1, "A1", "PROJECT PROGRESS REPORT")
	_ = f.SetCellStyle(sheet1, "A1", "A1", styleTitle)
	_ = f.SetCellValue(sheet1, "A2", fmt.Sprintf("%s — %s", order.NamaProject, order.NamaCustomer))

	// Write metadata box
	// Row 1
	_ = f.SetCellValue(sheet1, "A4", "Tanggal Kontrak")
	_ = f.SetCellValue(sheet1, "B4", tanggalKontrakStr)
	_ = f.SetCellValue(sheet1, "C4", "Customer")
	_ = f.SetCellValue(sheet1, "D4", order.NamaCustomer)
	_ = f.SetCellValue(sheet1, "E4", "Target Progres 50%")
	_ = f.SetCellValue(sheet1, "F4", target50Str)

	// Row 2
	_ = f.SetCellValue(sheet1, "A5", "Approval Gambar Kerja")
	_ = f.SetCellValue(sheet1, "B5", approvalGambarKerjaStr)
	_ = f.SetCellValue(sheet1, "E5", "Target Progres 80%")
	_ = f.SetCellValue(sheet1, "F5", target80Str)

	// Row 3
	_ = f.SetCellValue(sheet1, "A6", "Jangka Waktu")
	_ = f.SetCellValue(sheet1, "B6", jangkaWaktuStr)
	_ = f.SetCellValue(sheet1, "E6", "Target Progres 100%")
	_ = f.SetCellValue(sheet1, "F6", target100Str)

	// Merge Customer Box vertically
	_ = f.MergeCell(sheet1, "D4", "D6")

	// Set metadata styles
	for _, cell := range []string{"A4", "A5", "A6", "C4", "E4", "E5", "E6"} {
		_ = f.SetCellStyle(sheet1, cell, cell, styleLabel)
	}
	for _, cell := range []string{"B4", "B5", "B6", "F4", "F5", "F6"} {
		_ = f.SetCellStyle(sheet1, cell, cell, styleBodyC)
	}
	_ = f.SetCellStyle(sheet1, "D4", "D6", stylePinkBox)

	// Total progress info
	totalProgress := calculateProjectProgress(wp.Stages)
	_ = f.SetCellValue(sheet1, "A7", fmt.Sprintf("Total Progress: %.2f%%   |   Hari ini: %s", totalProgress, time.Now().Format("02/01/2006")))
	_ = f.SetCellStyle(sheet1, "A7", "A7", styleLabel)

	// Table Headers
	startRow := 9
	headers := []string{"NO", "JENIS PEKERJAAN", "QTY", "SATUAN", "MATERIAL", "BOBOT"}
	for colIdx, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+colIdx, startRow)
		_ = f.SetCellValue(sheet1, cell, h)
		_ = f.SetCellStyle(sheet1, cell, cell, styleHeader)
	}

	numStages := len(stageMasters)
	for mIdx, m := range stageMasters {
		colChar := 'G' + mIdx
		cellCol := fmt.Sprintf("%c%d", colChar, startRow)
		_ = f.SetCellValue(sheet1, cellCol, fmt.Sprintf("%s\n%.0f%%", m.Name, m.Percentage))
		_ = f.SetCellStyle(sheet1, cellCol, cellCol, styleHeader)
	}

	// Group rooms
	groupedRooms := make(map[string][]entity.InputItemRoom)
	var roomOrder []string
	seenRooms := make(map[string]bool)

	for _, r := range inputItem.Rooms {
		if !seenRooms[r.NamaRuangan] {
			seenRooms[r.NamaRuangan] = true
			roomOrder = append(roomOrder, r.NamaRuangan)
		}
		groupedRooms[r.NamaRuangan] = append(groupedRooms[r.NamaRuangan], r)
	}

	curRow := startRow + 1
	noCount := 1

	for _, rName := range roomOrder {
		// Room row
		endColChar := 'A' + len(headers) + numStages - 1
		_ = f.MergeCell(sheet1, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%c%d", endColChar, curRow))
		_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", curRow), strings.ToUpper(rName))
		_ = f.SetCellStyle(sheet1, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%c%d", endColChar, curRow), styleRoomRow)
		_ = f.SetRowHeight(sheet1, curRow, 22)
		curRow++

		for _, room := range groupedRooms[rName] {
			prodID := uint(0)
			if room.ProdukID != nil {
				prodID = *room.ProdukID
			}
			key := fmt.Sprintf("%s-%d-%.1f-%.1f-%.1f", room.NamaRuangan, prodID, room.Panjang, room.Lebar, room.Tinggi)
			hargaTotal := rabRoomPrices[key]

			bobot := 0.0
			if hasRAB && rab.GrandTotal > 0 {
				bobot = (hargaTotal / rab.GrandTotal) * 100
			} else {
				bobot = 100.0 / float64(len(inputItem.Rooms))
			}

			prodName := "Kustom"
			if room.Produk != nil {
				prodName = room.Produk.NamaProduk
			}

			materialStr := ""
			for mIdx, bb := range room.BahanBakus {
				if bb.BahanBaku != nil {
					if mIdx > 0 {
						materialStr += ", "
					}
					materialStr += bb.BahanBaku.NamaBahanBaku
				}
			}

			_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", curRow), noCount)
			_ = f.SetCellValue(sheet1, fmt.Sprintf("B%d", curRow), prodName)
			_ = f.SetCellValue(sheet1, fmt.Sprintf("C%d", curRow), room.Qty)
			_ = f.SetCellValue(sheet1, fmt.Sprintf("D%d", curRow), "Unit")
			_ = f.SetCellValue(sheet1, fmt.Sprintf("E%d", curRow), materialStr)
			_ = f.SetCellValue(sheet1, fmt.Sprintf("F%d", curRow), fmt.Sprintf("%.2f%%", bobot))

			for colIdx := 0; colIdx < 6; colIdx++ {
				cell := fmt.Sprintf("%c%d", 'A'+colIdx, curRow)
				if colIdx == 1 || colIdx == 4 {
					_ = f.SetCellStyle(sheet1, cell, cell, styleBody)
				} else if colIdx == 5 {
					_ = f.SetCellStyle(sheet1, cell, cell, styleBodyC) // right/center
				} else {
					_ = f.SetCellStyle(sheet1, cell, cell, styleBodyC)
				}
			}

			for mIdx, m := range stageMasters {
				colChar := 'G' + mIdx
				cellCol := fmt.Sprintf("%c%d", colChar, curRow)
				
				valStr := ""
				stage, stageExists := productStagesMap[room.ID][m.ID]
				if stageExists && stage.Status == "completed" {
					valStr = "✓"
				}
				
				_ = f.SetCellValue(sheet1, cellCol, valStr)
				_ = f.SetCellStyle(sheet1, cellCol, cellCol, styleBodyC)
			}

			_ = f.SetRowHeight(sheet1, curRow, 20)
			curRow++
			noCount++
		}
	}

	// Set column widths
	colWidths := map[string]float64{
		"A": 6,
		"B": 24,
		"C": 8,
		"D": 10,
		"E": 36,
		"F": 12,
	}
	for col, w := range colWidths {
		_ = f.SetColWidth(sheet1, col, col, w)
	}
	for mIdx := range stageMasters {
		colChar := fmt.Sprintf("%c", 'G'+mIdx)
		_ = f.SetColWidth(sheet1, colChar, colChar, 11)
	}

	// Sheet 2: Dokumentasi Foto
	sheet2 := "Dokumentasi Foto"
	_, _ = f.NewSheet(sheet2)

	_ = f.SetCellValue(sheet2, "A1", "DOKUMENTASI FOTO PROGRESS PRODUKSI")
	_ = f.SetCellStyle(sheet2, "A1", "A1", styleTitle)
	_ = f.SetCellValue(sheet2, "A2", fmt.Sprintf("Project: %s", order.NamaProject))

	// Fixed info headers: A-G, then Foto 1, Foto 2, ... dynamically
	infoHeaders := []string{"NO", "RUANGAN", "PRODUK", "TAHAPAN", "PIC SELESAI", "TANGGAL SELESAI", "CATATAN"}
	for colIdx, h := range infoHeaders {
		cell := fmt.Sprintf("%s%d", colIndexToName(colIdx), 4)
		_ = f.SetCellValue(sheet2, cell, h)
		_ = f.SetCellStyle(sheet2, cell, cell, styleHeader)
	}

	photoRow := 5
	photoIndex := 1
	maxPhotoCols := 0 // track max number of photo columns needed

	// Photo cell size (pixels). Excel column width unit ≈ 7px per unit, row height 1pt ≈ 1.33px
	// We target ~100x80px per photo cell
	const photoColWidthUnits = 15.0 // ~105px wide per photo column
	const photoRowHeightPt = 75.0   // 75pt ≈ 100px tall
	printTrue := true

	for _, rName := range roomOrder {
		for _, room := range groupedRooms[rName] {
			prodName := "Kustom"
			if room.Produk != nil {
				prodName = room.Produk.NamaProduk
			}

			for _, m := range stageMasters {
				stage, exists := productStagesMap[room.ID][m.ID]
				if !exists || stage.Photos == "" {
					continue
				}
				photosList := strings.Split(stage.Photos, ",")
				if len(photosList) == 0 || photosList[0] == "" {
					continue
				}

				timeStr := ""
				if stage.CompletedAt != nil {
					timeStr = stage.CompletedAt.Format("02/01/2006 15:04")
				}

				_ = f.SetRowHeight(sheet2, photoRow, photoRowHeightPt)

				// Info cells A-G
				infoValues := []interface{}{photoIndex, rName, prodName, m.Name, stage.CompletedBy, timeStr, stage.Notes}
				for colIdx, val := range infoValues {
					cell := fmt.Sprintf("%s%d", colIndexToName(colIdx), photoRow)
					_ = f.SetCellValue(sheet2, cell, val)
					if colIdx == 1 || colIdx == 2 || colIdx == 6 {
						_ = f.SetCellStyle(sheet2, cell, cell, styleBody)
					} else {
						_ = f.SetCellStyle(sheet2, cell, cell, styleBodyC)
					}
				}

				// Photos berderet mulai kolom H (index 7)
				validPhotoCount := 0
				for _, photoPath := range photosList {
					photoPath = strings.TrimSpace(photoPath)
					if photoPath == "" {
						continue
					}
					fullImgPath := s.resolvePhotoPath(photoPath)
					photoCellCol := 7 + validPhotoCount // 7 = H
					photoCell := fmt.Sprintf("%s%d", colIndexToName(photoCellCol), photoRow)
					_ = f.SetCellStyle(sheet2, photoCell, photoCell, styleBodyC)
					if _, errStat := os.Stat(fullImgPath); errStat == nil {
						_ = f.AddPicture(sheet2, photoCell, fullImgPath, &excelize.GraphicOptions{
							PrintObject:     &printTrue,
							LockAspectRatio: true,
							OffsetX:         3,
							OffsetY:         3,
							ScaleX:          0.13,
							ScaleY:          0.13,
							Positioning:     "oneCell",
						})
					} else {
						_ = f.SetCellValue(sheet2, photoCell, "N/A")
					}
					validPhotoCount++
				}

				if validPhotoCount > maxPhotoCols {
					maxPhotoCols = validPhotoCount
				}

				photoRow++
				photoIndex++
			}
		}
	}

	// Header for photo columns (Foto 1, Foto 2, ...)
	for i := 0; i < maxPhotoCols; i++ {
		colName := colIndexToName(7 + i)
		cell := fmt.Sprintf("%s4", colName)
		label := fmt.Sprintf("FOTO %d", i+1)
		_ = f.SetCellValue(sheet2, cell, label)
		_ = f.SetCellStyle(sheet2, cell, cell, styleHeader)
	}

	// Column widths for Sheet 2
	infoWidths := map[string]float64{
		"A": 6,
		"B": 18,
		"C": 24,
		"D": 15,
		"E": 18,
		"F": 18,
		"G": 28,
	}
	for col, w := range infoWidths {
		_ = f.SetColWidth(sheet2, col, col, w)
	}
	// Set photo column widths
	for i := 0; i < maxPhotoCols; i++ {
		colName := colIndexToName(7 + i)
		_ = f.SetColWidth(sheet2, colName, colName, photoColWidthUnits)
	}


	buf, errXls := f.WriteToBuffer()
	if errXls != nil {
		return nil, "", errXls
	}

	filename := fmt.Sprintf("Progress_Report_%s.xlsx", order.NomorOrder)
	return buf.Bytes(), filename, nil
}

// colIndexToName converts a 0-based column index to an Excel column name (A, B, ..., Z, AA, AB, ...)
func colIndexToName(idx int) string {
	name := ""
	for idx >= 0 {
		name = string(rune('A'+idx%26)) + name
		idx = idx/26 - 1
	}
	return name
}

// --- Defect Management ---

func defectToResponse(d entity.WorkplanDefect) dto.WorkplanDefectResponse {
	resp := dto.WorkplanDefectResponse{
		ID:              d.ID,
		WorkplanStageID: d.WorkplanStageID,
		Description:     d.Description,
		Photos:          d.Photos,
		Status:          d.Status,
		FixDescription:  d.FixDescription,
		FixPhotos:       d.FixPhotos,
		ReportedBy:      d.ReportedBy,
		FixedBy:         d.FixedBy,
		ReviewedBy:      d.ReviewedBy,
		ReviewNotes:     d.ReviewNotes,
		ReportedAt:      d.ReportedAt,
		FixedAt:         d.FixedAt,
		ReviewedAt:      d.ReviewedAt,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
	if d.WorkplanStage != nil {
		if d.WorkplanStage.StageMaster != nil {
			resp.StageName = d.WorkplanStage.StageMaster.Name
		}
		// Room/product name enrichment happens in the caller using raw query
	}
	return resp
}

func (s *workplanService) ReportDefect(ctx context.Context, stageID uint, req dto.ReportDefectRequest, userEmail string) (*dto.WorkplanDefectResponse, error) {
	// Verify stage exists
	var stage entity.WorkplanStage
	if err := s.db.WithContext(ctx).First(&stage, stageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tahapan tidak ditemukan")
		}
		return nil, err
	}

	// Check no existing active defect for this stage
	var existing entity.WorkplanDefect
	if err := s.db.WithContext(ctx).Where("workplan_stage_id = ? AND status IN ?", stageID, []string{"reported", "fix_submitted"}).First(&existing).Error; err == nil {
		return nil, errors.New("sudah ada laporan defect aktif pada tahapan ini")
	}

	now := time.Now()
	defect := entity.WorkplanDefect{
		WorkplanStageID: stageID,
		Description:     req.Description,
		Photos:          strings.Join(req.Photos, ","),
		Status:          "reported",
		ReportedBy:      userEmail,
		ReportedAt:      &now,
	}

	if err := s.db.WithContext(ctx).Create(&defect).Error; err != nil {
		return nil, fmt.Errorf("gagal menyimpan laporan defect: %w", err)
	}

	// Reload with relations
	s.db.WithContext(ctx).Preload("WorkplanStage.StageMaster").First(&defect, defect.ID)
	resp := defectToResponse(defect)
	return &resp, nil
}

func (s *workplanService) SubmitDefectFix(ctx context.Context, defectID uint, req dto.SubmitDefectFixRequest, userEmail string) (*dto.WorkplanDefectResponse, error) {
	var defect entity.WorkplanDefect
	if err := s.db.WithContext(ctx).Preload("WorkplanStage.StageMaster").First(&defect, defectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("defect tidak ditemukan")
		}
		return nil, err
	}

	if defect.Status != "reported" && defect.Status != "rejected" {
		return nil, errors.New("hanya defect yang berstatus 'reported' atau 'rejected' yang dapat diupload perbaikannya")
	}

	now := time.Now()
	defect.Status = "fix_submitted"
	defect.FixDescription = req.FixDescription
	defect.FixPhotos = strings.Join(req.FixPhotos, ",")
	defect.FixedBy = userEmail
	defect.FixedAt = &now

	if err := s.db.WithContext(ctx).Save(&defect).Error; err != nil {
		return nil, fmt.Errorf("gagal menyimpan perbaikan defect: %w", err)
	}

	resp := defectToResponse(defect)
	return &resp, nil
}

func (s *workplanService) ReviewDefect(ctx context.Context, defectID uint, req dto.ReviewDefectRequest, userEmail string) (*dto.WorkplanDefectResponse, error) {
	var defect entity.WorkplanDefect
	if err := s.db.WithContext(ctx).Preload("WorkplanStage.StageMaster").First(&defect, defectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("defect tidak ditemukan")
		}
		return nil, err
	}

	if defect.Status != "fix_submitted" {
		return nil, errors.New("hanya defect yang berstatus 'fix_submitted' yang dapat di-review")
	}

	now := time.Now()
	defect.ReviewedBy = userEmail
	defect.ReviewNotes = req.ReviewNotes
	defect.ReviewedAt = &now

	if req.Action == "accept" {
		defect.Status = "accepted"
	} else {
		defect.Status = "rejected"
	}

	if err := s.db.WithContext(ctx).Save(&defect).Error; err != nil {
		return nil, fmt.Errorf("gagal menyimpan keputusan review defect: %w", err)
	}

	resp := defectToResponse(defect)
	return &resp, nil
}

func (s *workplanService) GetDefectsByWorkplan(ctx context.Context, wpID uint) ([]dto.WorkplanDefectResponse, error) {
	// Get all stage IDs for this workplan
	var stageIDs []uint
	if err := s.db.WithContext(ctx).Model(&entity.WorkplanStage{}).Where("workplan_id = ?", wpID).Pluck("id", &stageIDs).Error; err != nil {
		return nil, err
	}

	if len(stageIDs) == 0 {
		return []dto.WorkplanDefectResponse{}, nil
	}

	var defects []entity.WorkplanDefect
	if err := s.db.WithContext(ctx).
		Preload("WorkplanStage.StageMaster").
		Preload("WorkplanStage.InputItemRoom.Produk").
		Where("workplan_stage_id IN ?", stageIDs).
		Order("created_at DESC").
		Find(&defects).Error; err != nil {
		return nil, err
	}

	result := make([]dto.WorkplanDefectResponse, 0, len(defects))
	for _, d := range defects {
		r := defectToResponse(d)
		// Enrich room/product name via InputItemRoom relation
		if d.WorkplanStage != nil {
			if d.WorkplanStage.InputItemRoom != nil {
				room := d.WorkplanStage.InputItemRoom
				r.RoomName = room.NamaRuangan
				if room.Produk != nil {
					r.ProductName = room.Produk.NamaProduk
				}
			}
			if d.WorkplanStage.StageMaster != nil {
				r.StageName = d.WorkplanStage.StageMaster.Name
			}
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *workplanService) SubmitBast(ctx context.Context, id uint, req dto.SubmitBastRequest, email string) (*dto.WorkplanResponse, error) {
	wp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cek semua tahapan selesai
	for _, stage := range wp.Stages {
		if stage.Status != "completed" {
			return nil, errors.New("tidak dapat memproses BAST: masih terdapat tahapan pengerjaan yang belum selesai")
		}
	}

	// Cek jika ada defect aktif
	var activeDefectsCount int64
	var stageIDs []uint
	for _, stage := range wp.Stages {
		stageIDs = append(stageIDs, stage.ID)
	}
	if len(stageIDs) > 0 {
		s.db.Model(&entity.WorkplanDefect{}).Where("workplan_stage_id IN ? AND status IN ('reported', 'fix_submitted')", stageIDs).Count(&activeDefectsCount)
		if activeDefectsCount > 0 {
			return nil, errors.New("tidak dapat memproses BAST: masih terdapat defect/cacat yang aktif")
		}
	}

	now := time.Now()
	wp.BastPhoto = req.BastPhoto
	wp.BastGeneratedAt = &now
	wp.BastGeneratedBy = email

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(wp).Error; err != nil {
			return err
		}
		// Update order tahapan_proyek to selesai and payment_status to lunas
		if err := tx.Model(&entity.Order{}).Where("id = ?", wp.OrderID).Updates(map[string]interface{}{
			"payment_status": "lunas",
		}).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	// Transition stage to selesai and log transition
	_ = s.logTaskSvc.TransitionStage(ctx, wp.OrderID, "selesai", email)

	return s.GetByID(ctx, wp.ID)
}

func (s *workplanService) GenerateBastPDF(ctx context.Context, id uint) ([]byte, string, error) {
	wp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	var order entity.Order
	if errOrder := s.db.WithContext(ctx).Preload("Contracts").First(&order, wp.OrderID).Error; errOrder != nil {
		return nil, "", errOrder
	}

	// Cek semua termin lunas
	var invoices []entity.Invoice
	if errInv := s.db.Where("order_id = ?", wp.OrderID).Find(&invoices).Error; errInv == nil {
		if len(invoices) == 0 {
			return nil, "", errors.New("tagihan belum dibuat, tidak dapat men-generate BAST")
		}
		for _, inv := range invoices {
			if strings.ToLower(inv.Keterangan) == "bast" {
				continue
			}
			if inv.Status != "terbayar" {
				return nil, "", errors.New("pembayaran belum dilunasi, silakan selesaikan seluruh tagihan terlebih dahulu")
			}
		}
	} else {
		return nil, "", errInv
	}

	// Update order status to selesai and payment_status to lunas
	if errOrderUpdate := s.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", wp.OrderID).Updates(map[string]interface{}{
		"tahapan_proyek": "selesai",
		"payment_status": "lunas",
	}).Error; errOrderUpdate != nil {
		s.logger.Error("failed to update order status during BAST generation", zap.Error(errOrderUpdate))
	}

	// Initialize PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	// Fetch Company Profile
	cp := entity.GetCompanyProfile(s.db, order.CompanyID)

	logoFile := ""
	if cp.Logo != "" {
		logoFile = filepath.Join(s.uploadDir, filepath.Base(cp.Logo))
		if _, err := os.Stat(logoFile); os.IsNotExist(err) {
			logoFile = ""
		}
	}

	if logoFile != "" {
		pdf.Image(logoFile, 20, 15, 0, 15, false, "", 0, "")
		pdf.SetLeftMargin(38)
		pdf.SetX(38)
		pdf.SetY(15)
	} else {
		pdf.SetY(15)
	}

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 128, 128) // Teal
	pdf.CellFormat(0, 7, cp.Name, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 4, "Premium Interior Design & Architecture Services", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, fmt.Sprintf("Email: %s | Phone: %s", cp.Email, cp.Phone), "", 1, "L", false, 0, "")

	// Reset margins
	pdf.SetLeftMargin(20)
	pdf.SetX(20)
	pdf.SetY(32)

	pdf.SetDrawColor(0, 128, 128)
	pdf.SetLineWidth(0.8)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	// Document Title
	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 6, "BERITA ACARA SERAH TERIMA (BAST)", "", 1, "C", false, 0, "")

	now := time.Now()
	docNo := fmt.Sprintf("No: %d/BAST/NUG/%s/%d", wp.ID, now.Format("01"), now.Year())
	pdf.SetFont("Arial", "I", 9.5)
	pdf.CellFormat(0, 5, docNo, "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Body Text - Opening
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(60, 60, 60)
	
	// Translate day to indonesian
	dayEN := now.Format("Monday")
	dayID := "Senin"
	switch dayEN {
	case "Monday":
		dayID = "Senin"
	case "Tuesday":
		dayID = "Selasa"
	case "Wednesday":
		dayID = "Rabu"
	case "Thursday":
		dayID = "Kamis"
	case "Friday":
		dayID = "Jumat"
	case "Saturday":
		dayID = "Sabtu"
	case "Sunday":
		dayID = "Minggu"
	}
	introText := fmt.Sprintf("Pada hari ini, %s tanggal %s, telah dilakukan serah terima hasil pekerjaan oleh dan di antara pihak-pihak:",
		dayID, now.Format("02 January 2006"))
	pdf.MultiCell(0, 5, introText, "", "L", false)
	pdf.Ln(4)

	// Party 1 - First Party (Arsiflow)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 5, "1. PIHAK PERTAMA (Penyedia Jasa):", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(25)
	pdf.CellFormat(0, 5, "Nama Perusahaan : "+cp.Name, "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, 5, "Alamat               : "+cp.Address, "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, 5, "Telepon              : "+cp.Phone, "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Party 2 - Second Party (Customer)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 5, "2. PIHAK KEDUA (Pemilik Proyek / Customer):", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(25)
	pdf.CellFormat(0, 5, "Nama Customer  : "+order.NamaCustomer, "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, 5, "Alamat               : "+order.Alamat, "", 1, "L", false, 0, "")
	pdf.SetX(25)
	pdf.CellFormat(0, 5, "Telepon              : "+order.TeleponCustomer, "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// Statement
	statement := fmt.Sprintf("Kedua belah pihak dengan ini menyatakan bahwa PIHAK PERTAMA telah menyelesaikan seluruh tahapan pengerjaan interior untuk proyek \"%s\" (Order No: %s) dengan persentase kemajuan 100%% (Selesai).\n\nPIHAK KEDUA menyatakan telah memeriksa dan menerima seluruh hasil pengerjaan tersebut dalam kondisi baik, tanpa cacat, dan sesuai spesifikasi yang disepakati bersama. Terhitung sejak tanggal penandatanganan berita acara ini, tanggung jawab pemeliharaan hasil pengerjaan beralih sepenuhnya kepada PIHAK KEDUA.",
		order.NamaProject, order.NomorOrder)
	pdf.MultiCell(0, 5, statement, "", "J", false)
	pdf.Ln(10)

	// Date and Signatures
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Jakarta, %s", now.Format("02 January 2006")), "", 1, "R", false, 0, "")
	pdf.Ln(4)

	// Two column signatures
	pdf.SetFont("Arial", "B", 10)
	
	// Left: Pihak Pertama
	pdf.SetX(20)
	pdf.CellFormat(80, 5, "PIHAK PERTAMA,", "", 0, "C", false, 0, "")
	
	// Right: Pihak Kedua
	pdf.SetX(110)
	pdf.CellFormat(80, 5, "PIHAK KEDUA,", "", 1, "C", false, 0, "")
	
	pdf.SetFont("Arial", "", 9)
	pdf.SetX(20)
	pdf.CellFormat(80, 5, cp.Name, "", 0, "C", false, 0, "")
	pdf.SetX(110)
	pdf.CellFormat(80, 5, "Pemilik Proyek / Customer", "", 1, "C", false, 0, "")
	
	// Spacing for signatures
	pdf.Ln(20)
	
	pdf.SetFont("Arial", "BU", 10)
	pdf.SetX(20)
	pdf.CellFormat(80, 5, "   ( Arsiflow Team )   ", "", 0, "C", false, 0, "")
	pdf.SetX(110)
	pdf.CellFormat(80, 5, fmt.Sprintf("   ( %s )   ", order.NamaCustomer), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("BAST_%s_%d.pdf", strings.ReplaceAll(order.NomorOrder, "/", "_"), wp.ID)
	return buf.Bytes(), filename, nil
}

