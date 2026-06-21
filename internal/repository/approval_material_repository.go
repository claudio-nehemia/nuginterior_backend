package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ApprovalMaterialRepository interface {
	FindAll(ctx context.Context) ([]entity.ApprovalMaterial, error)
	FindByID(ctx context.Context, id uint) (*entity.ApprovalMaterial, error)
	FindByOrderID(ctx context.Context, orderID uint) (*entity.ApprovalMaterial, error)
	Create(ctx context.Context, am *entity.ApprovalMaterial) error
	Update(ctx context.Context, am *entity.ApprovalMaterial) error
	Delete(ctx context.Context, id uint) error

	SaveItem(ctx context.Context, item *entity.ApprovalMaterialItem) error
	FindItemByID(ctx context.Context, itemID uint) (*entity.ApprovalMaterialItem, error)
	UpdateOrderStage(ctx context.Context, orderID uint, stage string) error
}

type approvalMaterialRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewApprovalMaterialRepository(db *gorm.DB, logger *zap.Logger) ApprovalMaterialRepository {
	return &approvalMaterialRepository{db: db, logger: logger}
}

func (r *approvalMaterialRepository) FindAll(ctx context.Context) ([]entity.ApprovalMaterial, error) {
	var list []entity.ApprovalMaterial
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Items").
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *approvalMaterialRepository) FindByID(ctx context.Context, id uint) (*entity.ApprovalMaterial, error) {
	var am entity.ApprovalMaterial
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Items").
		First(&am, id).Error
	if err != nil {
		return nil, err
	}
	return &am, nil
}

func (r *approvalMaterialRepository) FindByOrderID(ctx context.Context, orderID uint) (*entity.ApprovalMaterial, error) {
	var am entity.ApprovalMaterial
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Items").
		Where("order_id = ?", orderID).
		First(&am).Error
	if err != nil {
		return nil, err
	}
	return &am, nil
}

func (r *approvalMaterialRepository) Create(ctx context.Context, am *entity.ApprovalMaterial) error {
	return r.db.WithContext(ctx).Create(am).Error
}

func (r *approvalMaterialRepository) Update(ctx context.Context, am *entity.ApprovalMaterial) error {
	return r.db.WithContext(ctx).Save(am).Error
}

func (r *approvalMaterialRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.ApprovalMaterial{}, id).Error
}

func (r *approvalMaterialRepository) SaveItem(ctx context.Context, item *entity.ApprovalMaterialItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *approvalMaterialRepository) FindItemByID(ctx context.Context, itemID uint) (*entity.ApprovalMaterialItem, error) {
	var item entity.ApprovalMaterialItem
	err := r.db.WithContext(ctx).First(&item, itemID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *approvalMaterialRepository) UpdateOrderStage(ctx context.Context, orderID uint, stage string) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Update("tahapan_proyek", stage).Error
}
