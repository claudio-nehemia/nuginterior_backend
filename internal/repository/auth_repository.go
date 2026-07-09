package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthRepository handles auth-related database operations.
type AuthRepository interface {
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	FindUserByID(ctx context.Context, id uint) (*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) error
	FindCompanyByID(ctx context.Context, id uint) (*entity.Company, error)
	CreateCompanyAndUser(ctx context.Context, company *entity.Company, user *entity.User) error
}

type authRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewAuthRepository(db *gorm.DB, logger *zap.Logger) AuthRepository {
	return &authRepository{db: db, logger: logger}
}

func (r *authRepository) FindUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Role.Divisi").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) FindUserByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Company").
		Preload("Role").
		Preload("Role.Divisi").
		Preload("Role.Permissions").
		First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) CreateUser(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *authRepository) FindCompanyByID(ctx context.Context, id uint) (*entity.Company, error) {
	var company entity.Company
	err := r.db.WithContext(ctx).First(&company, id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *authRepository) CreateCompanyAndUser(ctx context.Context, company *entity.Company, user *entity.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get default_active_days from global settings (company_id = 1)
		var val string
		err := tx.Table("settings").Where("company_id = ? AND key = ?", 1, "default_active_days").Pluck("value", &val).Error
		days := 4 // Default fallback
		if err == nil && val != "" {
			if parsedDays, err := strconv.Atoi(val); err == nil {
				days = parsedDays
			}
		}

		expiredAt := time.Now().AddDate(0, 0, days)
		company.ExpiredAt = &expiredAt

		if err := tx.Create(company).Error; err != nil {
			return err
		}
		user.CompanyID = company.ID
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return nil
	})
}
