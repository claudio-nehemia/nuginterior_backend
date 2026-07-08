package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrderRepository handles order-related database operations.
type OrderRepository interface {
	FindAll(ctx context.Context, search, status string) ([]entity.Order, error)
	FindAllFiltered(ctx context.Context, params map[string]string) ([]entity.Order, error)
	FindByID(ctx context.Context, id uint) (*entity.Order, error)
	Create(ctx context.Context, order *entity.Order) error
	Update(ctx context.Context, order *entity.Order) error
	Delete(ctx context.Context, id uint) error
	GenerateNomorOrder(ctx context.Context) (string, error)
	SyncTeams(ctx context.Context, orderID uint, userIDs []uint) error
	GetTeams(ctx context.Context, orderID uint) ([]entity.OrderTeam, error)
}

type orderRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewOrderRepository(db *gorm.DB, logger *zap.Logger) OrderRepository {
	return &orderRepository{db: db, logger: logger}
}

func (r *orderRepository) FindAll(ctx context.Context, search, status string) ([]entity.Order, error) {
	var orders []entity.Order
	query := r.db.WithContext(ctx).Scopes(database.CompanyScope(ctx))
	if search != "" {
		q := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(nama_project) LIKE ? OR LOWER(nama_customer) LIKE ? OR LOWER(nama_perusahaan) LIKE ? OR LOWER(nomor_order) LIKE ?",
			q, q, q, q,
		)
	}
	if status != "" {
		query = query.Where("project_status = ?", status)
	}
	err := query.Preload("Teams").Order("id DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindByID(ctx context.Context, id uint) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		Preload("Termin").
		Preload("PIC").
		Preload("Teams.User").
		Preload("Teams.User.Role").
		Preload("Surveys.Order").
		Preload("Surveys.Surveyor").
		Preload("Surveys.SurveyPengukuran.JenisPengukuran").
		Preload("Moodboards.Files").
		Preload("Moodboards.Estimasi.Files").
		Preload("Moodboards.CommitmentFee").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		companyID := database.GetContextCompanyID(ctx)
		if order.CompanyID == 0 && companyID > 0 {
			order.CompanyID = companyID
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		var tanggal *time.Time
		if order.TanggalSurvey != "" {
			t, err := time.Parse("2006-01-02", order.TanggalSurvey)
			if err == nil {
				tanggal = &t
			}
		}

		survey := entity.Survey{
			OrderID:       order.ID,
			TanggalSurvey: tanggal,
			Lokasi:        order.Alamat,
			Status:        "pending",
		}

		if err := tx.Create(&survey).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		var tanggal *time.Time
		if order.TanggalSurvey != "" {
			t, err := time.Parse("2006-01-02", order.TanggalSurvey)
			if err == nil {
				tanggal = &t
			}
		}

		if err := tx.Model(&entity.Survey{}).Where("order_id = ?", order.ID).Update("tanggal_survey", tanggal).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *orderRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete associated surveys
		if err := tx.Where("order_id = ?", id).Delete(&entity.Survey{}).Error; err != nil {
			return err
		}
		// Delete associated teams
		if err := tx.Where("order_id = ?", id).Delete(&entity.OrderTeam{}).Error; err != nil {
			return err
		}
		// Delete order
		if err := tx.Delete(&entity.Order{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *orderRepository) GenerateNomorOrder(ctx context.Context) (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("ORD-%s-", now.Format("200601"))
	var count int64
	r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("nomor_order LIKE ?", prefix+"%").Count(&count)
	return fmt.Sprintf("%s%03d", prefix, count+1), nil
}

func (r *orderRepository) SyncTeams(ctx context.Context, orderID uint, userIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", orderID).Delete(&entity.OrderTeam{}).Error; err != nil {
			return err
		}
		for _, uid := range userIDs {
			team := entity.OrderTeam{OrderID: orderID, UserID: uid}
			if err := tx.Create(&team).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) GetTeams(ctx context.Context, orderID uint) ([]entity.OrderTeam, error) {
	var teams []entity.OrderTeam
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Preload("User").
		Preload("User.Role").
		Find(&teams).Error
	return teams, err
}

func (r *orderRepository) FindAllFiltered(ctx context.Context, params map[string]string) ([]entity.Order, error) {
	var orders []entity.Order
	query := r.db.WithContext(ctx).Scopes(database.CompanyScope(ctx))

	if search, ok := params["search"]; ok && search != "" {
		q := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(nama_project) LIKE ? OR LOWER(nama_customer) LIKE ? OR LOWER(nama_perusahaan) LIKE ? OR LOWER(nomor_order) LIKE ?",
			q, q, q, q,
		)
	}

	if status, ok := params["status"]; ok && status != "" && status != "all" {
		query = query.Where("project_status = ?", status)
	}

	if tahapan, ok := params["tahapan_proyek"]; ok && tahapan != "" && tahapan != "all" {
		query = query.Where("tahapan_proyek = ?", tahapan)
	}

	if payment, ok := params["payment_status"]; ok && payment != "" && payment != "all" {
		query = query.Where("payment_status = ?", payment)
	}

	if priority, ok := params["priority_level"]; ok && priority != "" && priority != "all" {
		query = query.Where("priority_level = ?", priority)
	}

	if jenis, ok := params["jenis_interior"]; ok && jenis != "" && jenis != "all" {
		query = query.Where("jenis_interior = ?", jenis)
	}

	if startDate, ok := params["start_date"]; ok && startDate != "" {
		query = query.Where("tanggal_masuk_customer >= ?", startDate)
	}

	if endDate, ok := params["end_date"]; ok && endDate != "" {
		query = query.Where("tanggal_masuk_customer <= ?", endDate)
	}

	err := query.Preload("Teams").Preload("PIC").Preload("Contracts").Preload("Contracts.RAB").Order("id DESC").Find(&orders).Error
	return orders, err
}
