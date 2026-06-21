package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProjectLogTaskService interface {
	GetAll(ctx context.Context) ([]dto.ProjectLogTaskResponse, error)
	TransitionStage(ctx context.Context, orderID uint, toStage string, userEmail string) error
	RecordTouch(ctx context.Context, orderID uint, stage string, userEmail string) error
}

type projectLogTaskService struct {
	repo            repository.ProjectLogTaskRepository
	orderRepo       repository.OrderRepository
	settingRepo     repository.SettingRepository
	logger          *zap.Logger
	notificationSvc NotificationService
}

func NewProjectLogTaskService(
	repo repository.ProjectLogTaskRepository,
	orderRepo repository.OrderRepository,
	settingRepo repository.SettingRepository,
	logger *zap.Logger,
	notificationSvc NotificationService,
) ProjectLogTaskService {
	return &projectLogTaskService{
		repo:            repo,
		orderRepo:       orderRepo,
		settingRepo:     settingRepo,
		logger:          logger,
		notificationSvc: notificationSvc,
	}
}

func (s *projectLogTaskService) GetAll(ctx context.Context) ([]dto.ProjectLogTaskResponse, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	res := []dto.ProjectLogTaskResponse{}
	now := time.Now()

	for _, item := range list {
		nomorOrder := ""
		namaProject := ""
		namaCustomer := ""
		if item.Order != nil {
			nomorOrder = item.Order.NomorOrder
			namaProject = item.Order.NamaProject
			namaCustomer = item.Order.NamaCustomer
		}

		deadlineTime := item.CreatedAt.AddDate(0, 0, item.DeadlineDays)

		isLate := false
		lateDays := 0
		var compareTime time.Time
		if item.CompletedAt != nil {
			compareTime = *item.CompletedAt
		} else {
			compareTime = now
		}

		if compareTime.After(deadlineTime) && item.DeadlineDays > 0 {
			isLate = true
			diff := compareTime.Sub(deadlineTime)
			lateDays = int(diff.Hours() / 24)
			if lateDays == 0 && diff > 0 {
				lateDays = 1
			}
		}

		durationToTouch := "-"
		if item.TouchedAt != nil {
			durationToTouch = dto.FormatDurationIndonesian(item.TouchedAt.Sub(item.CreatedAt))
		}

		durationToComplete := "-"
		if item.CompletedAt != nil {
			startForComplete := item.CreatedAt
			if item.TouchedAt != nil {
				startForComplete = *item.TouchedAt
			}
			durationToComplete = dto.FormatDurationIndonesian(item.CompletedAt.Sub(startForComplete))
		}

		resp := dto.ProjectLogTaskResponse{
			ID:                 item.ID,
			OrderID:            item.OrderID,
			NomorOrder:         nomorOrder,
			NamaProject:        namaProject,
			NamaCustomer:       namaCustomer,
			Stage:              item.Stage,
			StageLabel:         getStageLabel(item.Stage),
			CreatedAt:          item.CreatedAt,
			TouchedAt:          item.TouchedAt,
			TouchedBy:          item.TouchedBy,
			CompletedAt:        item.CompletedAt,
			CompletedBy:        item.CompletedBy,
			DeadlineDays:       item.DeadlineDays,
			DeadlineTime:       deadlineTime,
			IsLate:             isLate,
			LateDays:           lateDays,
			DurationToTouch:    durationToTouch,
			DurationToComplete: durationToComplete,
		}
		res = append(res, resp)
	}

	return res, nil
}

func (s *projectLogTaskService) TransitionStage(ctx context.Context, orderID uint, toStage string, userEmail string) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	fromStage := order.TahapanProyek
	if fromStage == toStage {
		return nil
	}

	now := time.Now()

	// Update order stage
	order.TahapanProyek = toStage
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return err
	}

	// Complete active log of previous stage
	if fromStage != "" && fromStage != "not_start" {
		activeLog, err := s.repo.FindActiveByOrderIDAndStage(ctx, orderID, fromStage)
		if err == nil && activeLog != nil {
			activeLog.CompletedAt = &now
			resolvedEmail := userEmail
			if resolvedEmail == "" {
				resolvedEmail = GetUserEmailFromContext(ctx)
			}
			activeLog.CompletedBy = resolvedEmail
			_ = s.repo.Update(ctx, activeLog)
		}
	}

	// Create new stage log
	deadlineDays := 0
	if toStage != "selesai" && toStage != "batal" {
		setting, errSetting := s.settingRepo.FindByKey(ctx, "deadline_stage_"+toStage)
		if errSetting == nil && setting != nil {
			deadlineDays, _ = strconv.Atoi(setting.Value)
		} else {
			deadlineDays = 3 // default fallback
		}
	}

	newLog := &entity.ProjectLogTask{
		OrderID:      orderID,
		Stage:        toStage,
		CreatedAt:    now,
		DeadlineDays: deadlineDays,
	}

	if err := s.repo.Create(ctx, newLog); err != nil {
		return err
	}

	// Trigger notifications
	var eventName string
	var title string
	var message string
	var link string = fmt.Sprintf("/dashboard/order/%d", orderID)

	switch toStage {
	case "moodboard":
		eventName = "moodboard"
		title = "Mulai Tahap Moodboard"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Moodboard.", order.NamaProject, order.NomorOrder)
	case "estimasi":
		eventName = "estimasi"
		title = "Mulai Tahap Estimasi"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Estimasi.", order.NamaProject, order.NomorOrder)
	case "cm_fee":
		eventName = "commitment_fee"
		title = "Mulai Tahap Commitment Fee"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Commitment Fee.", order.NamaProject, order.NomorOrder)
	case "desain_final":
		eventName = "desain_final"
		title = "Mulai Tahap Desain Final"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Desain Final.", order.NamaProject, order.NomorOrder)
	case "input_item":
		eventName = "input_item"
		title = "Mulai Tahap Input Item"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Input Item.", order.NamaProject, order.NomorOrder)
	case "rab":
		eventName = "rab"
		title = "Mulai Tahap RAB"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan RAB.", order.NamaProject, order.NomorOrder)
	case "kontrak":
		eventName = "kontrak"
		title = "Mulai Tahap Kontrak"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Kontrak.", order.NamaProject, order.NomorOrder)
	case "invoice":
		eventName = "invoice"
		title = "Mulai Tahap Invoice"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Invoice.", order.NamaProject, order.NomorOrder)
	case "survey_ulang":
		eventName = "setup_survey_ulang"
		title = "Mulai Tahap Setup Survey Ulang"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Setup Survey Ulang.", order.NamaProject, order.NomorOrder)
	case "gambar_kerja":
		eventName = "gambar_kerja"
		title = "Mulai Tahap Gambar Kerja"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Gambar Kerja.", order.NamaProject, order.NomorOrder)
	case "approval_material", "workplan":
		eventName = "approval_material"
		title = "Mulai Tahap Approval Material / Workplan"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan %s.", order.NamaProject, order.NomorOrder, getStageLabel(toStage))
	case "operations":
		eventName = "project_management"
		title = "Mulai Tahap Project Management"
		message = fmt.Sprintf("Proyek %s (Order %s) kini berada pada tahapan Operations / Project Management.", order.NamaProject, order.NomorOrder)
	}

	if eventName != "" {
		_ = s.notificationSvc.SendNotification(ctx, orderID, eventName, title, message, link)
	}

	return nil
}

func GetUserEmailFromContext(ctx context.Context) string {
	if email, ok := ctx.Value(constants.ContextKeyUserEmail).(string); ok {
		return email
	}
	return "system"
}

func (s *projectLogTaskService) RecordTouch(ctx context.Context, orderID uint, stage string, userEmail string) error {
	resolvedEmail := userEmail
	if resolvedEmail == "" {
		resolvedEmail = GetUserEmailFromContext(ctx)
	}

	activeLog, err := s.repo.FindActiveByOrderIDAndStage(ctx, orderID, stage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			deadlineDays := 3
			setting, errSetting := s.settingRepo.FindByKey(ctx, "deadline_stage_"+stage)
			if errSetting == nil && setting != nil {
				deadlineDays, _ = strconv.Atoi(setting.Value)
			}
			now := time.Now()
			activeLog = &entity.ProjectLogTask{
				OrderID:      orderID,
				Stage:        stage,
				CreatedAt:    now,
				TouchedAt:    &now,
				TouchedBy:    resolvedEmail,
				DeadlineDays: deadlineDays,
			}
			return s.repo.Create(ctx, activeLog)
		}
		return err
	}

	if activeLog.TouchedAt == nil {
		now := time.Now()
		activeLog.TouchedAt = &now
		activeLog.TouchedBy = resolvedEmail
		return s.repo.Update(ctx, activeLog)
	}

	return nil
}

func getStageLabel(stage string) string {
	labels := map[string]string{
		"survey":            "Survey",
		"moodboard":         "Moodboard",
		"estimasi":          "Estimasi",
		"cm_fee":            "CM Fee",
		"desain_final":      "Desain Final",
		"input_item":        "Input Item",
		"rab":               "RAB",
		"kontrak":           "Kontrak",
		"invoice":           "Invoice",
		"survey_ulang":      "Survey Ulang",
		"gambar_kerja":      "Gambar Kerja",
		"approval_material": "Approval Material",
		"workplan":          "Workplan",
		"operations":        "Operations",
		"selesai":           "Selesai",
		"batal":             "Batal",
	}
	if label, ok := labels[stage]; ok {
		return label
	}
	return stage
}
