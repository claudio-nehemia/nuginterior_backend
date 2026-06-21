package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type JenisPengukuranRepository interface {
	FindAll(ctx context.Context) ([]entity.JenisPengukuran, error)
	FindByID(ctx context.Context, id uint) (*entity.JenisPengukuran, error)
	Create(ctx context.Context, jp *entity.JenisPengukuran) error
	Update(ctx context.Context, jp *entity.JenisPengukuran) error
	SoftDelete(ctx context.Context, id uint) error
}

type jenisPengukuranRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewJenisPengukuranRepository(db *gorm.DB, logger *zap.Logger) JenisPengukuranRepository {
	return &jenisPengukuranRepository{db: db, logger: logger}
}

func (r *jenisPengukuranRepository) FindAll(ctx context.Context) ([]entity.JenisPengukuran, error) {
	var list []entity.JenisPengukuran
	err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *jenisPengukuranRepository) FindByID(ctx context.Context, id uint) (*entity.JenisPengukuran, error) {
	var jp entity.JenisPengukuran
	err := r.db.WithContext(ctx).First(&jp, id).Error
	if err != nil {
		return nil, err
	}
	return &jp, nil
}

func (r *jenisPengukuranRepository) Create(ctx context.Context, jp *entity.JenisPengukuran) error {
	return r.db.WithContext(ctx).Create(jp).Error
}

func (r *jenisPengukuranRepository) Update(ctx context.Context, jp *entity.JenisPengukuran) error {
	return r.db.WithContext(ctx).Save(jp).Error
}

func (r *jenisPengukuranRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.JenisPengukuran{}, id).Error
}
