package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WorkplanRepository interface {
	FindAll(ctx context.Context) ([]entity.Workplan, error)
	FindByID(ctx context.Context, id uint) (*entity.Workplan, error)
	FindByOrderID(ctx context.Context, orderID uint) (*entity.Workplan, error)
	Create(ctx context.Context, wp *entity.Workplan) error
	Update(ctx context.Context, wp *entity.Workplan) error
	Delete(ctx context.Context, id uint) error

	SaveStage(ctx context.Context, stage *entity.WorkplanStage) error
	DeleteStage(ctx context.Context, id uint) error
	FindStageByID(ctx context.Context, id uint) (*entity.WorkplanStage, error)
	FindStagesByWorkplanID(ctx context.Context, workplanID uint) ([]entity.WorkplanStage, error)
	GetStageMasters(ctx context.Context) ([]entity.WorkplanStageMaster, error)
	UpdateOrderStage(ctx context.Context, orderID uint, stage string) error
}

type workplanRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewWorkplanRepository(db *gorm.DB, logger *zap.Logger) WorkplanRepository {
	return &workplanRepository{db: db, logger: logger}
}

func (r *workplanRepository) FindAll(ctx context.Context) ([]entity.Workplan, error) {
	var list []entity.Workplan
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Stages.StageMaster").
		Preload("Stages.InputItemRoom.Produk").
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *workplanRepository) FindByID(ctx context.Context, id uint) (*entity.Workplan, error) {
	var wp entity.Workplan
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Stages.StageMaster").
		Preload("Stages.InputItemRoom.Produk").
		First(&wp, id).Error
	if err != nil {
		return nil, err
	}
	return &wp, nil
}

func (r *workplanRepository) FindByOrderID(ctx context.Context, orderID uint) (*entity.Workplan, error) {
	var wp entity.Workplan
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Stages.StageMaster").
		Preload("Stages.InputItemRoom.Produk").
		Where("order_id = ?", orderID).
		First(&wp).Error
	if err != nil {
		return nil, err
	}
	return &wp, nil
}

func (r *workplanRepository) Create(ctx context.Context, wp *entity.Workplan) error {
	return r.db.WithContext(ctx).Create(wp).Error
}

func (r *workplanRepository) Update(ctx context.Context, wp *entity.Workplan) error {
	return r.db.WithContext(ctx).Save(wp).Error
}

func (r *workplanRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Workplan{}, id).Error
}

func (r *workplanRepository) SaveStage(ctx context.Context, stage *entity.WorkplanStage) error {
	return r.db.WithContext(ctx).Save(stage).Error
}

func (r *workplanRepository) DeleteStage(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.WorkplanStage{}, id).Error
}

func (r *workplanRepository) FindStageByID(ctx context.Context, id uint) (*entity.WorkplanStage, error) {
	var stage entity.WorkplanStage
	err := r.db.WithContext(ctx).First(&stage, id).Error
	if err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *workplanRepository) FindStagesByWorkplanID(ctx context.Context, workplanID uint) ([]entity.WorkplanStage, error) {
	var stages []entity.WorkplanStage
	err := r.db.WithContext(ctx).
		Preload("StageMaster").
		Preload("InputItemRoom.Produk").
		Where("workplan_id = ?", workplanID).
		Find(&stages).Error
	return stages, err
}

func (r *workplanRepository) GetStageMasters(ctx context.Context) ([]entity.WorkplanStageMaster, error) {
	var masters []entity.WorkplanStageMaster
	err := r.db.WithContext(ctx).Order("sort_order ASC").Find(&masters).Error
	return masters, err
}

func (r *workplanRepository) UpdateOrderStage(ctx context.Context, orderID uint, stage string) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Update("tahapan_proyek", stage).Error
}
