package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationService interface {
	SendNotification(ctx context.Context, orderID uint, event string, defaultTitle string, defaultMessage string, link string) error
	GetNotificationsForUser(ctx context.Context, userID uint) ([]dto.NotificationResponse, error)
	MarkAsRead(ctx context.Context, id uint, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	CheckDeadlines(ctx context.Context) error
}

type notificationService struct {
	repo        repository.NotificationRepository
	settingRepo repository.SettingRepository
	db          *gorm.DB
	logger      *zap.Logger
}

func NewNotificationService(
	repo repository.NotificationRepository,
	settingRepo repository.SettingRepository,
	db *gorm.DB,
	logger *zap.Logger,
) NotificationService {
	return &notificationService{
		repo:        repo,
		settingRepo: settingRepo,
		db:          db,
		logger:      logger,
	}
}

type RoleConfig struct {
	RoleName string `json:"role_name"`
	TeamOnly bool   `json:"team_only"`
}

type RuleConfig struct {
	Name  string       `json:"name"`
	Roles []RoleConfig `json:"roles"`
}

type NotificationSettings struct {
	AlwaysNotifiedRoles []string              `json:"always_notified_roles"`
	Rules               map[string]RuleConfig `json:"rules"`
}

func (s *notificationService) SendNotification(ctx context.Context, orderID uint, event string, defaultTitle string, defaultMessage string, link string) error {
	s.logger.Info("Sending notification", zap.Uint("order_id", orderID), zap.String("event", event))

	// 1. Fetch notification settings
	setting, err := s.settingRepo.FindByKey(ctx, "notification_settings")
	if err != nil {
		s.logger.Error("Failed to fetch notification settings", zap.Error(err))
		return err
	}

	var config NotificationSettings
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		s.logger.Error("Failed to unmarshal notification settings", zap.Error(err))
		return err
	}

	// 2. Determine target user IDs
	recipients := make(map[uint]bool)

	// Fetch order if orderID is valid
	var order entity.Order
	if orderID > 0 {
		if err := s.db.Preload("PIC").First(&order, orderID).Error; err != nil {
			s.logger.Error("Failed to fetch order for notification", zap.Uint("order_id", orderID), zap.Error(err))
		}
	}

	// Process event rules
	if rule, ok := config.Rules[event]; ok {
		for _, rConfig := range rule.Roles {
			if rConfig.TeamOnly && orderID > 0 {
				// Determine which team to look up: Survey Ulang Team or Order Team
				isSurveyUlangEvent := event == "upload_survey_ulang" ||
					event == "gambar_kerja" ||
					event == "approval_material" ||
					event == "project_management"

				var teamIDs []uint
				if isSurveyUlangEvent {
					// Get latest survey to find survey_ulang_team_ids
					var latestSurvey entity.Survey
					if err := s.db.Where("order_id = ?", orderID).Order("created_at desc").First(&latestSurvey).Error; err == nil {
						if len(latestSurvey.SurveyUlangTeamIDs) > 0 {
							_ = json.Unmarshal(latestSurvey.SurveyUlangTeamIDs, &teamIDs)
						}
					}
				}

				if len(teamIDs) > 0 {
					// Retrieve users with matching role from survey ulang team
					var userIDs []uint
					s.db.Table("users").
						Joins("JOIN roles ON roles.id = users.role_id").
						Where("users.id IN ? AND roles.nama_role = ?", teamIDs, rConfig.RoleName).
						Pluck("users.id", &userIDs)
					for _, uid := range userIDs {
						recipients[uid] = true
					}
				} else {
					// Retrieve users with matching role from order_teams
					var userIDs []uint
					s.db.Table("order_teams").
						Joins("JOIN users ON users.id = order_teams.user_id").
						Joins("JOIN roles ON roles.id = users.role_id").
						Where("order_teams.order_id = ? AND roles.nama_role = ?", orderID, rConfig.RoleName).
						Pluck("users.id", &userIDs)
					for _, uid := range userIDs {
						recipients[uid] = true
					}
				}
			} else {
				// Notify all users with this role
				var userIDs []uint
				s.db.Table("users").
					Joins("JOIN roles ON roles.id = users.role_id").
					Where("roles.nama_role = ?", rConfig.RoleName).
					Pluck("users.id", &userIDs)
				for _, uid := range userIDs {
					recipients[uid] = true
				}
			}
		}
	}

	// 3. Process Global Subscribers/Always Notified Roles
	if orderID > 0 {
		for _, alwaysRole := range config.AlwaysNotifiedRoles {
			if alwaysRole == "Kepala Marketing" {
				// Kepala Marketing always gets notifications from initial team setup until end
				var userIDs []uint
				s.db.Table("order_teams").
					Joins("JOIN users ON users.id = order_teams.user_id").
					Joins("JOIN roles ON roles.id = users.role_id").
					Where("order_teams.order_id = ? AND roles.nama_role = ?", orderID, "Kepala Marketing").
					Pluck("users.id", &userIDs)
				for _, uid := range userIDs {
					recipients[uid] = true
				}
			} else if alwaysRole == "Project Manager" {
				// Project Manager always gets notifications after survey ulang setup
				var latestSurvey entity.Survey
				if err := s.db.Where("order_id = ?", orderID).Order("created_at desc").First(&latestSurvey).Error; err == nil {
					if len(latestSurvey.SurveyUlangTeamIDs) > 0 {
						var teamIDs []uint
						_ = json.Unmarshal(latestSurvey.SurveyUlangTeamIDs, &teamIDs)
						if len(teamIDs) > 0 {
							var pmUserIDs []uint
							s.db.Table("users").
								Joins("JOIN roles ON roles.id = users.role_id").
								Where("users.id IN ? AND roles.nama_role = ?", teamIDs, "Project Manager").
								Pluck("users.id", &pmUserIDs)
							for _, uid := range pmUserIDs {
								recipients[uid] = true
							}
						}
					}
				}
			}
		}
	}

	// 4. Save notification records
	for userID := range recipients {
		var oID *uint
		if orderID > 0 {
			oID = &orderID
		}
		notif := &entity.Notification{
			UserID:    userID,
			OrderID:   oID,
			Title:     defaultTitle,
			Message:   defaultMessage,
			Link:      link,
			IsRead:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.repo.Create(ctx, notif); err != nil {
			s.logger.Error("Failed to save notification record", zap.Uint("user_id", userID), zap.Error(err))
		}
	}

	return nil
}

func (s *notificationService) GetNotificationsForUser(ctx context.Context, userID uint) ([]dto.NotificationResponse, error) {
	list, err := s.repo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.NotificationResponse, len(list))
	for i, n := range list {
		nomorOrder := ""
		namaProject := ""
		if n.Order != nil {
			nomorOrder = n.Order.NomorOrder
			namaProject = n.Order.NamaProject
		}
		resp[i] = dto.NotificationResponse{
			ID:          n.ID,
			OrderID:     n.OrderID,
			NomorOrder:  nomorOrder,
			NamaProject: namaProject,
			Title:       n.Title,
			Message:     n.Message,
			Link:        n.Link,
			IsRead:      n.IsRead,
			CreatedAt:   n.CreatedAt,
		}
	}
	return resp, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uint, userID uint) error {
	notif, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if notif.UserID != userID {
		return fmt.Errorf("unauthorized to update this notification")
	}
	notif.IsRead = true
	return s.repo.Update(ctx, notif)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *notificationService) CheckDeadlines(ctx context.Context) error {
	s.logger.Info("Checking project stage deadlines for warnings...")

	// 1. Fetch active tasks
	var activeLogs []entity.ProjectLogTask
	err := s.db.Preload("Order").Where("completed_at IS NULL AND deadline_days > 0").Find(&activeLogs).Error
	if err != nil {
		s.logger.Error("Failed to fetch active log tasks for deadlines check", zap.Error(err))
		return err
	}

	now := time.Now()

	for _, logTask := range activeLogs {
		if logTask.Order == nil {
			continue
		}

		deadlineTime := logTask.CreatedAt.AddDate(0, 0, logTask.DeadlineDays)
		remaining := deadlineTime.Sub(now)

		// Check if H-1 (less than or equal to 24 hours, and not already expired)
		if remaining <= 24*time.Hour && remaining > 0 {
			stageLabel := getStageLabel(logTask.Stage)
			title := "Peringatan H-1: Batas Waktu Tahap " + stageLabel
			message := fmt.Sprintf("Proyek %s (Order %s) akan mencapai batas waktu untuk tahapan %s dalam 24 jam.",
				logTask.Order.NamaProject, logTask.Order.NomorOrder, stageLabel)
			link := fmt.Sprintf("/dashboard/order/%d", logTask.OrderID)

			// Resolve recipients for this active stage
			recipients := s.resolveStageRecipients(ctx, logTask.OrderID, logTask.Stage)

			for userID := range recipients {
				// Check if duplicate warning already sent
				var exists int64
				s.db.Model(&entity.Notification{}).
					Where("user_id = ? AND order_id = ? AND title = ?", userID, logTask.OrderID, title).
					Count(&exists)

				if exists == 0 {
					notif := &entity.Notification{
						UserID:    userID,
						OrderID:   &logTask.OrderID,
						Title:     title,
						Message:   message,
						Link:      link,
						IsRead:    false,
						CreatedAt: now,
						UpdatedAt: now,
					}
					if err := s.repo.Create(ctx, notif); err != nil {
						s.logger.Error("Failed to save deadline warning notification", zap.Uint("user_id", userID), zap.Error(err))
					}
				}
			}
		}
	}

	return nil
}

func (s *notificationService) resolveStageRecipients(ctx context.Context, orderID uint, stage string) map[uint]bool {
	recipients := make(map[uint]bool)

	// Fetch settings
	setting, err := s.settingRepo.FindByKey(ctx, "notification_settings")
	if err != nil {
		return recipients
	}

	var config NotificationSettings
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		return recipients
	}

	// For warnings, use the roles defined for that stage
	if rule, ok := config.Rules[stage]; ok {
		for _, rConfig := range rule.Roles {
			if rConfig.TeamOnly && orderID > 0 {
				isSurveyUlangEvent := stage == "upload_survey_ulang" ||
					stage == "gambar_kerja" ||
					stage == "approval_material" ||
					stage == "project_management"

				var teamIDs []uint
				if isSurveyUlangEvent {
					var latestSurvey entity.Survey
					if err := s.db.Where("order_id = ?", orderID).Order("created_at desc").First(&latestSurvey).Error; err == nil {
						if len(latestSurvey.SurveyUlangTeamIDs) > 0 {
							_ = json.Unmarshal(latestSurvey.SurveyUlangTeamIDs, &teamIDs)
						}
					}
				}

				if len(teamIDs) > 0 {
					var userIDs []uint
					s.db.Table("users").
						Joins("JOIN roles ON roles.id = users.role_id").
						Where("users.id IN ? AND roles.nama_role = ?", teamIDs, rConfig.RoleName).
						Pluck("users.id", &userIDs)
					for _, uid := range userIDs {
						recipients[uid] = true
					}
				} else {
					var userIDs []uint
					s.db.Table("order_teams").
						Joins("JOIN users ON users.id = order_teams.user_id").
						Joins("JOIN roles ON roles.id = users.role_id").
						Where("order_teams.order_id = ? AND roles.nama_role = ?", orderID, rConfig.RoleName).
						Pluck("users.id", &userIDs)
					for _, uid := range userIDs {
						recipients[uid] = true
					}
				}
			} else {
				var userIDs []uint
				s.db.Table("users").
					Joins("JOIN roles ON roles.id = users.role_id").
					Where("roles.nama_role = ?", rConfig.RoleName).
					Pluck("users.id", &userIDs)
				for _, uid := range userIDs {
					recipients[uid] = true
				}
			}
		}
	}

	// Always add Kepala Marketing (from initial team) and Project Manager (from survey ulang team)
	for _, alwaysRole := range config.AlwaysNotifiedRoles {
		if alwaysRole == "Kepala Marketing" {
			var userIDs []uint
			s.db.Table("order_teams").
				Joins("JOIN users ON users.id = order_teams.user_id").
				Joins("JOIN roles ON roles.id = users.role_id").
				Where("order_teams.order_id = ? AND roles.nama_role = ?", orderID, "Kepala Marketing").
				Pluck("users.id", &userIDs)
			for _, uid := range userIDs {
				recipients[uid] = true
			}
		} else if alwaysRole == "Project Manager" {
			var latestSurvey entity.Survey
			if err := s.db.Where("order_id = ?", orderID).Order("created_at desc").First(&latestSurvey).Error; err == nil {
				if len(latestSurvey.SurveyUlangTeamIDs) > 0 {
					var teamIDs []uint
					_ = json.Unmarshal(latestSurvey.SurveyUlangTeamIDs, &teamIDs)
					if len(teamIDs) > 0 {
						var pmUserIDs []uint
						s.db.Table("users").
							Joins("JOIN roles ON roles.id = users.role_id").
							Where("users.id IN ? AND roles.nama_role = ?", teamIDs, "Project Manager").
							Pluck("users.id", &pmUserIDs)
						for _, uid := range pmUserIDs {
							recipients[uid] = true
						}
					}
				}
			}
		}
	}

	return recipients
}
