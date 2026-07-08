package repository

import (
	"context"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SurveyRepository handles survey-related database operations.
type SurveyRepository interface {
	FindAll(ctx context.Context) ([]entity.Survey, error)
	FindByID(ctx context.Context, id uint) (*entity.Survey, error)
	Create(ctx context.Context, survey *entity.Survey) error
	Update(ctx context.Context, survey *entity.Survey) error
	Delete(ctx context.Context, id uint) error
	SyncPengukuran(ctx context.Context, surveyID uint, items []entity.SurveyPengukuran) error
	FindUsersByIDs(ctx context.Context, ids []uint) ([]entity.User, error)
	UpdateOrderStage(ctx context.Context, orderID uint, stage string) error
	IsMarketingResponseEnabled(ctx context.Context) (bool, error)
}

type surveyRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewSurveyRepository(db *gorm.DB, logger *zap.Logger) SurveyRepository {
	return &surveyRepository{db: db, logger: logger}
}

func (r *surveyRepository) FindAll(ctx context.Context) ([]entity.Survey, error) {
	// 1. Fetch all orders (scoped to company)
	var orders []entity.Order
	if err := r.db.WithContext(ctx).Scopes(database.CompanyScope(ctx)).Find(&orders).Error; err != nil {
		return nil, err
	}

	// 2. Fetch existing surveys
	var surveys []entity.Survey
	if err := r.db.WithContext(ctx).Find(&surveys).Error; err != nil {
		return nil, err
	}

	// 3. Find which orders don't have a survey and create them
	surveyedOrders := make(map[uint]bool)
	for _, s := range surveys {
		surveyedOrders[s.OrderID] = true
	}

	for _, o := range orders {
		if !surveyedOrders[o.ID] {
			var tanggal *time.Time
			if o.TanggalSurvey != "" {
				t, err := time.Parse("2006-01-02", o.TanggalSurvey)
				if err == nil {
					tanggal = &t
				}
			}
			newSurvey := entity.Survey{
				OrderID:       o.ID,
				TanggalSurvey: tanggal,
				Lokasi:        o.Alamat,
				Status:        "pending",
			}
			if err := r.db.WithContext(ctx).Create(&newSurvey).Error; err != nil {
				r.logger.Error("failed to auto-create missing survey for order", zap.Uint("order_id", o.ID), zap.Error(err))
				continue
			}
		}
	}

	// 4. Query surveys with all preloads and proper ordering (scoped to company orders)
	var result []entity.Survey
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Order.Contracts").
		Preload("Order.Teams.User").
		Preload("Order.Teams.User.Role").
		Preload("Surveyor").
		Preload("SurveyPengukuran.JenisPengukuran").
		Order("id DESC").
		Find(&result).Error
	return result, err
}

func (r *surveyRepository) FindByID(ctx context.Context, id uint) (*entity.Survey, error) {
	var survey entity.Survey
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Order.Contracts").
		Preload("Order.Teams.User").
		Preload("Order.Teams.User.Role").
		Preload("Surveyor").
		Preload("SurveyPengukuran.JenisPengukuran").
		First(&survey, id).Error
	if err != nil {
		return nil, err
	}
	return &survey, nil
}

func (r *surveyRepository) Create(ctx context.Context, survey *entity.Survey) error {
	return r.db.WithContext(ctx).Create(survey).Error
}

func (r *surveyRepository) Update(ctx context.Context, survey *entity.Survey) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(survey).Error; err != nil {
			return err
		}

		var tglStr string
		if survey.TanggalSurvey != nil {
			tglStr = survey.TanggalSurvey.Format("2006-01-02")
		}

		if err := tx.Model(&entity.Order{}).Where("id = ?", survey.OrderID).Update("tanggal_survey", tglStr).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *surveyRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Survey{}, id).Error
}

func (r *surveyRepository) SyncPengukuran(ctx context.Context, surveyID uint, items []entity.SurveyPengukuran) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("survey_id = ?", surveyID).Delete(&entity.SurveyPengukuran{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].SurveyID = surveyID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *surveyRepository) FindUsersByIDs(ctx context.Context, ids []uint) ([]entity.User, error) {
	var users []entity.User
	if len(ids) == 0 {
		return users, nil
	}
	err := r.db.WithContext(ctx).Preload("Role").Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *surveyRepository) UpdateOrderStage(ctx context.Context, orderID uint, stage string) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Update("tahapan_proyek", stage).Error
}

func (r *surveyRepository) IsMarketingResponseEnabled(ctx context.Context) (bool, error) {
	var val string
	err := r.db.WithContext(ctx).Scopes(database.CompanyScope(ctx)).Model(&entity.Setting{}).Where("key = ?", "marketing_response_enabled").Pluck("value", &val).Error
	if err != nil {
		return true, err
	}
	return val == "true", nil
}
