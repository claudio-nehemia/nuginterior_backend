package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DivisiRepository handles divisi-related database operations.
type DivisiRepository interface {
	FindAll(ctx context.Context) ([]entity.Divisi, error)
	FindByID(ctx context.Context, id uint) (*entity.Divisi, error)
	Create(ctx context.Context, divisi *entity.Divisi) error
	Update(ctx context.Context, divisi *entity.Divisi) error
	Delete(ctx context.Context, id uint) error
}

type divisiRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewDivisiRepository(db *gorm.DB, logger *zap.Logger) DivisiRepository {
	return &divisiRepository{db: db, logger: logger}
}

func (r *divisiRepository) FindAll(ctx context.Context) ([]entity.Divisi, error) {
	var divisis []entity.Divisi
	err := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		Order("id ASC").Find(&divisis).Error
	return divisis, err
}

func (r *divisiRepository) FindByID(ctx context.Context, id uint) (*entity.Divisi, error) {
	var divisi entity.Divisi
	err := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		First(&divisi, id).Error
	if err != nil {
		return nil, err
	}
	return &divisi, nil
}

func (r *divisiRepository) Create(ctx context.Context, divisi *entity.Divisi) error {
	companyID, _ := ctx.Value(constants.ContextKeyCompanyID).(uint)
	if companyID == 0 {
		companyID = 1
	}
	divisi.CompanyID = companyID
	return r.db.WithContext(ctx).Create(divisi).Error
}

func (r *divisiRepository) Update(ctx context.Context, divisi *entity.Divisi) error {
	return r.db.WithContext(ctx).Save(divisi).Error
}

func (r *divisiRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Divisi{}, id).Error
}
