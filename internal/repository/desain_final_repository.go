package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DesainFinalRepository interface {
	FindAll(ctx context.Context) ([]entity.DesainFinal, error)
	FindByID(ctx context.Context, id uint) (*entity.DesainFinal, error)
	FindByOrderID(ctx context.Context, orderID uint) (*entity.DesainFinal, error)
	Create(ctx context.Context, df *entity.DesainFinal) error
	Update(ctx context.Context, df *entity.DesainFinal) error
	Delete(ctx context.Context, id uint) error

	FindFileByID(ctx context.Context, fileID uint) (*entity.DesainFinalFile, error)
	CreateFile(ctx context.Context, file *entity.DesainFinalFile) error
	UpdateFile(ctx context.Context, file *entity.DesainFinalFile) error
	DeleteFile(ctx context.Context, fileID uint) error

	UpdateOrderStageAndPayment(ctx context.Context, orderID uint, stage string, paymentStatus string) error
}

type desainFinalRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewDesainFinalRepository(db *gorm.DB, logger *zap.Logger) DesainFinalRepository {
	return &desainFinalRepository{db: db, logger: logger}
}

func (r *desainFinalRepository) FindAll(ctx context.Context) ([]entity.DesainFinal, error) {
	var list []entity.DesainFinal
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *desainFinalRepository) FindByID(ctx context.Context, id uint) (*entity.DesainFinal, error) {
	var df entity.DesainFinal
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		First(&df, id).Error
	if err != nil {
		return nil, err
	}
	return &df, nil
}

func (r *desainFinalRepository) FindByOrderID(ctx context.Context, orderID uint) (*entity.DesainFinal, error) {
	var df entity.DesainFinal
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		Where("order_id = ?", orderID).
		First(&df).Error
	if err != nil {
		return nil, err
	}
	return &df, nil
}

func (r *desainFinalRepository) Create(ctx context.Context, df *entity.DesainFinal) error {
	return r.db.WithContext(ctx).Create(df).Error
}

func (r *desainFinalRepository) Update(ctx context.Context, df *entity.DesainFinal) error {
	return r.db.WithContext(ctx).Save(df).Error
}

func (r *desainFinalRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.DesainFinal{}, id).Error
}

func (r *desainFinalRepository) FindFileByID(ctx context.Context, fileID uint) (*entity.DesainFinalFile, error) {
	var file entity.DesainFinalFile
	err := r.db.WithContext(ctx).First(&file, fileID).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *desainFinalRepository) CreateFile(ctx context.Context, file *entity.DesainFinalFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *desainFinalRepository) UpdateFile(ctx context.Context, file *entity.DesainFinalFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *desainFinalRepository) DeleteFile(ctx context.Context, fileID uint) error {
	return r.db.WithContext(ctx).Delete(&entity.DesainFinalFile{}, fileID).Error
}

func (r *desainFinalRepository) UpdateOrderStageAndPayment(ctx context.Context, orderID uint, stage string, paymentStatus string) error {
	updates := map[string]interface{}{}
	if stage != "" {
		updates["tahapan_proyek"] = stage
	}
	if paymentStatus != "" {
		updates["payment_status"] = paymentStatus
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Updates(updates).Error
}
