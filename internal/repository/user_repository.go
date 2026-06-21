package repository

import (
	"context"
	"strings"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRepository handles user-related database operations.
type UserRepository interface {
	FindAll(ctx context.Context, search string, roleID uint) ([]entity.User, error)
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
	ExistsByEmail(ctx context.Context, email string, excludeID ...uint) (bool, error)
}

type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &userRepository{db: db, logger: logger}
}

func (r *userRepository) FindAll(ctx context.Context, search string, roleID uint) ([]entity.User, error) {
	var users []entity.User
	query := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		Preload("Role").
		Preload("Role.Divisi")
	if search != "" {
		q := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", q, q)
	}
	if roleID > 0 {
		query = query.Where("role_id = ?", roleID)
	}
	err := query.Order("id ASC").Find(&users).Error
	return users, err
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		Preload("Role").
		Preload("Role.Divisi").
		First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string, excludeID ...uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email)
	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}
	err := query.Count(&count).Error
	return count > 0, err
}
