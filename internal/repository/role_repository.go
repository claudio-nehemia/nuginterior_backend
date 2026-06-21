package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleRepository handles role-related database operations.
type RoleRepository interface {
	FindAll(ctx context.Context) ([]entity.Role, error)
	FindByID(ctx context.Context, id uint) (*entity.Role, error)
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uint) error
	CountUsersByRoleID(ctx context.Context, roleID uint) (int64, error)
	CountPermissionsByRoleID(ctx context.Context, roleID uint) (int64, error)
	SyncPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
	GetPermissionsByRoleID(ctx context.Context, roleID uint) ([]entity.Permission, error)
}

type roleRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewRoleRepository(db *gorm.DB, logger *zap.Logger) RoleRepository {
	return &roleRepository{db: db, logger: logger}
}

func (r *roleRepository) FindAll(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		Preload("Divisi").
		Order("id ASC").
		Find(&roles).Error
	return roles, err
}

func (r *roleRepository) FindByID(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		Preload("Divisi").
		Preload("Users").
		Preload("Permissions").
		First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	companyID, _ := ctx.Value(constants.ContextKeyCompanyID).(uint)
	if companyID == 0 {
		companyID = 1
	}
	role.CompanyID = companyID
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

func (r *roleRepository) CountUsersByRoleID(ctx context.Context, roleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

func (r *roleRepository) CountPermissionsByRoleID(ctx context.Context, roleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.RolePermission{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

func (r *roleRepository) SyncPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing permissions
		if err := tx.Where("role_id = ?", roleID).Delete(&entity.RolePermission{}).Error; err != nil {
			return err
		}
		// Insert new permissions
		for _, permID := range permissionIDs {
			rp := entity.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRepository) GetPermissionsByRoleID(ctx context.Context, roleID uint) ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permission ON role_permission.permission_id = permissions.id").
		Where("role_permission.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}
