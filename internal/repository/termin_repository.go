package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TerminRepository interface {
	FindAll(ctx context.Context) ([]entity.Termin, error)
	FindByID(ctx context.Context, id uint) (*entity.Termin, error)
	Create(ctx context.Context, termin *entity.Termin) error
	Update(ctx context.Context, termin *entity.Termin) error
	Delete(ctx context.Context, id uint) error
}

type terminRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewTerminRepository(db *gorm.DB, logger *zap.Logger) TerminRepository {
	return &terminRepository{db: db, logger: logger}
}

func (r *terminRepository) FindAll(ctx context.Context) ([]entity.Termin, error) {
	var list []entity.Termin
	err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *terminRepository) FindByID(ctx context.Context, id uint) (*entity.Termin, error) {
	var t entity.Termin
	err := r.db.WithContext(ctx).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *terminRepository) Create(ctx context.Context, termin *entity.Termin) error {
	return r.db.WithContext(ctx).Create(termin).Error
}

func (r *terminRepository) Update(ctx context.Context, termin *entity.Termin) error {
	return r.db.WithContext(ctx).Save(termin).Error
}

func (r *terminRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Termin{}, id).Error
}
