package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionRepository handles permission-related database operations.
type PermissionRepository interface {
	FindAll(ctx context.Context) ([]entity.Permission, error)
	FindByID(ctx context.Context, id uint) (*entity.Permission, error)
	Upsert(ctx context.Context, perm *entity.Permission) error
}

type permissionRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewPermissionRepository(db *gorm.DB, logger *zap.Logger) PermissionRepository {
	return &permissionRepository{db: db, logger: logger}
}

func (r *permissionRepository) FindAll(ctx context.Context) ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.WithContext(ctx).Order("\"group\" ASC, name ASC").Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) FindByID(ctx context.Context, id uint) (*entity.Permission, error) {
	var perm entity.Permission
	err := r.db.WithContext(ctx).First(&perm, id).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *permissionRepository) Upsert(ctx context.Context, perm *entity.Permission) error {
	return r.db.WithContext(ctx).
		Where("name = ?", perm.Name).
		Assign(entity.Permission{
			DisplayName: perm.DisplayName,
			Group:       perm.Group,
		}).
		FirstOrCreate(perm).Error
}
