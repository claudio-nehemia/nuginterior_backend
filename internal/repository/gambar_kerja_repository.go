package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GambarKerjaRepository interface {
	FindAll(ctx context.Context) ([]entity.GambarKerja, error)
	FindByID(ctx context.Context, id uint) (*entity.GambarKerja, error)
	FindByOrderID(ctx context.Context, orderID uint) (*entity.GambarKerja, error)
	Create(ctx context.Context, gk *entity.GambarKerja) error
	Update(ctx context.Context, gk *entity.GambarKerja) error
	Delete(ctx context.Context, id uint) error

	FindFileByID(ctx context.Context, fileID uint) (*entity.GambarKerjaFile, error)
	CreateFile(ctx context.Context, file *entity.GambarKerjaFile) error
	UpdateFile(ctx context.Context, file *entity.GambarKerjaFile) error
	DeleteFile(ctx context.Context, fileID uint) error

	UpdateOrderStageAndPayment(ctx context.Context, orderID uint, stage string, paymentStatus string) error
}

type gambarKerjaRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewGambarKerjaRepository(db *gorm.DB, logger *zap.Logger) GambarKerjaRepository {
	return &gambarKerjaRepository{db: db, logger: logger}
}

func (r *gambarKerjaRepository) FindAll(ctx context.Context) ([]entity.GambarKerja, error) {
	var list []entity.GambarKerja
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *gambarKerjaRepository) FindByID(ctx context.Context, id uint) (*entity.GambarKerja, error) {
	var gk entity.GambarKerja
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		First(&gk, id).Error
	if err != nil {
		return nil, err
	}
	return &gk, nil
}

func (r *gambarKerjaRepository) FindByOrderID(ctx context.Context, orderID uint) (*entity.GambarKerja, error) {
	var gk entity.GambarKerja
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		Where("order_id = ?", orderID).
		First(&gk).Error
	if err != nil {
		return nil, err
	}
	return &gk, nil
}

func (r *gambarKerjaRepository) Create(ctx context.Context, gk *entity.GambarKerja) error {
	return r.db.WithContext(ctx).Create(gk).Error
}

func (r *gambarKerjaRepository) Update(ctx context.Context, gk *entity.GambarKerja) error {
	return r.db.WithContext(ctx).Save(gk).Error
}

func (r *gambarKerjaRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.GambarKerja{}, id).Error
}

func (r *gambarKerjaRepository) FindFileByID(ctx context.Context, fileID uint) (*entity.GambarKerjaFile, error) {
	var file entity.GambarKerjaFile
	err := r.db.WithContext(ctx).First(&file, fileID).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *gambarKerjaRepository) CreateFile(ctx context.Context, file *entity.GambarKerjaFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *gambarKerjaRepository) UpdateFile(ctx context.Context, file *entity.GambarKerjaFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *gambarKerjaRepository) DeleteFile(ctx context.Context, fileID uint) error {
	return r.db.WithContext(ctx).Delete(&entity.GambarKerjaFile{}, fileID).Error
}

func (r *gambarKerjaRepository) UpdateOrderStageAndPayment(ctx context.Context, orderID uint, stage string, paymentStatus string) error {
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
