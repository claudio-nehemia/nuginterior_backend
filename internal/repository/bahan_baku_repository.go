package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BahanBakuRepository handles bahan_baku-related database operations.
type BahanBakuRepository interface {
	FindAll(ctx context.Context) ([]entity.BahanBaku, error)
	FindByID(ctx context.Context, id uint) (*entity.BahanBaku, error)
	Create(ctx context.Context, bb *entity.BahanBaku) error
	Update(ctx context.Context, bb *entity.BahanBaku) error
	Delete(ctx context.Context, id uint) error
}

type bahanBakuRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewBahanBakuRepository(db *gorm.DB, logger *zap.Logger) BahanBakuRepository {
	return &bahanBakuRepository{db: db, logger: logger}
}

func (r *bahanBakuRepository) FindAll(ctx context.Context) ([]entity.BahanBaku, error) {
	var list []entity.BahanBaku
	err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *bahanBakuRepository) FindByID(ctx context.Context, id uint) (*entity.BahanBaku, error) {
	var bb entity.BahanBaku
	err := r.db.WithContext(ctx).First(&bb, id).Error
	if err != nil {
		return nil, err
	}
	return &bb, nil
}

func (r *bahanBakuRepository) Create(ctx context.Context, bb *entity.BahanBaku) error {
	return r.db.WithContext(ctx).Create(bb).Error
}

func (r *bahanBakuRepository) Update(ctx context.Context, bb *entity.BahanBaku) error {
	return r.db.WithContext(ctx).Save(bb).Error
}

func (r *bahanBakuRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.BahanBaku{}, id).Error
}
