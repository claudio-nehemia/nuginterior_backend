package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SettingRepository handles setting-related database operations.
type SettingRepository interface {
	FindAll(ctx context.Context) ([]entity.Setting, error)
	FindByKey(ctx context.Context, key string) (*entity.Setting, error)
	Upsert(ctx context.Context, setting *entity.Setting) error
}

type settingRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewSettingRepository(db *gorm.DB, logger *zap.Logger) SettingRepository {
	return &settingRepository{db: db, logger: logger}
}

func (r *settingRepository) FindAll(ctx context.Context) ([]entity.Setting, error) {
	var list []entity.Setting
	err := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *settingRepository) FindByKey(ctx context.Context, key string) (*entity.Setting, error) {
	var setting entity.Setting
	err := r.db.WithContext(ctx).
		Scopes(database.CompanyScope(ctx)).
		Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *settingRepository) Upsert(ctx context.Context, setting *entity.Setting) error {
	companyID, _ := ctx.Value(constants.ContextKeyCompanyID).(uint)
	if companyID == 0 {
		companyID = 1
	}
	setting.CompanyID = companyID

	err := r.db.WithContext(ctx).
		Where("company_id = ? AND key = ?", companyID, setting.Key).
		Assign(entity.Setting{Value: setting.Value, Description: setting.Description}).
		FirstOrCreate(setting).Error
	if err != nil {
		return err
	}

	// Sync with companies table if it matches one of company fields
	var companyField string
	switch setting.Key {
	case "company_name":
		companyField = "name"
	case "company_director":
		companyField = "director_name"
	case "company_logo":
		companyField = "logo"
	case "company_address":
		companyField = "address"
	case "company_bank_name":
		companyField = "bank_name"
	case "company_bank_account":
		companyField = "bank_account"
	case "company_bank_holder":
		companyField = "bank_holder"
	case "company_email":
		companyField = "email"
	case "company_phone":
		companyField = "phone"
	}

	if companyField != "" {
		err = r.db.WithContext(ctx).Table("companies").Where("id = ?", companyID).Update(companyField, setting.Value).Error
		if err != nil {
			r.logger.Error("Failed to sync setting to companies table", zap.Uint("company_id", companyID), zap.String("key", setting.Key), zap.Error(err))
		}
	}

	return nil
}
