package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProjectLogTaskRepository interface {
	FindAll(ctx context.Context) ([]entity.ProjectLogTask, error)
	FindActiveByOrderIDAndStage(ctx context.Context, orderID uint, stage string) (*entity.ProjectLogTask, error)
	Create(ctx context.Context, log *entity.ProjectLogTask) error
	Update(ctx context.Context, log *entity.ProjectLogTask) error
}

type projectLogTaskRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewProjectLogTaskRepository(db *gorm.DB, logger *zap.Logger) ProjectLogTaskRepository {
	return &projectLogTaskRepository{db: db, logger: logger}
}

func (r *projectLogTaskRepository) FindAll(ctx context.Context) ([]entity.ProjectLogTask, error) {
	var list []entity.ProjectLogTask
	err := r.db.WithContext(ctx).
		Preload("Order").
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *projectLogTaskRepository) FindActiveByOrderIDAndStage(ctx context.Context, orderID uint, stage string) (*entity.ProjectLogTask, error) {
	var log entity.ProjectLogTask
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND stage = ? AND completed_at IS NULL", orderID, stage).
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *projectLogTaskRepository) Create(ctx context.Context, log *entity.ProjectLogTask) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *projectLogTaskRepository) Update(ctx context.Context, log *entity.ProjectLogTask) error {
	return r.db.WithContext(ctx).Save(log).Error
}
