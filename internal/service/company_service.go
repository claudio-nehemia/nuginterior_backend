package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CompanyService interface {
	GetAll(ctx context.Context) ([]dto.CompanyResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.CompanyResponse, error)
	Update(ctx context.Context, id uint, req dto.CompanyUpdateRequest) (*dto.CompanyResponse, error)
	Verify(ctx context.Context, id uint) error
	Reject(ctx context.Context, id uint) error
}

type companyService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewCompanyService(db *gorm.DB, logger *zap.Logger) CompanyService {
	return &companyService{db: db, logger: logger}
}

func (s *companyService) GetAll(ctx context.Context) ([]dto.CompanyResponse, error) {
	var companies []entity.Company
	err := s.db.WithContext(ctx).Order("id asc").Find(&companies).Error
	if err != nil {
		return nil, err
	}

	res := make([]dto.CompanyResponse, len(companies))
	for i, c := range companies {
		var adminEmail string
		s.db.WithContext(ctx).Model(&entity.User{}).Where("company_id = ?", c.ID).Order("id asc").Limit(1).Pluck("email", &adminEmail)
		res[i] = toCompanyResponse(c, adminEmail)
	}
	return res, nil
}

func (s *companyService) GetByID(ctx context.Context, id uint) (*dto.CompanyResponse, error) {
	var company entity.Company
	err := s.db.WithContext(ctx).First(&company, id).Error
	if err != nil {
		return nil, err
	}
	var adminEmail string
	s.db.WithContext(ctx).Model(&entity.User{}).Where("company_id = ?", company.ID).Order("id asc").Limit(1).Pluck("email", &adminEmail)
	res := toCompanyResponse(company, adminEmail)
	return &res, nil
}

func (s *companyService) Update(ctx context.Context, id uint, req dto.CompanyUpdateRequest) (*dto.CompanyResponse, error) {
	var company entity.Company
	if err := s.db.WithContext(ctx).First(&company, id).Error; err != nil {
		return nil, err
	}

	company.Name = req.Name
	company.DirectorName = req.DirectorName
	company.CeoNik = req.CeoNik
	company.Nib = req.Nib
	company.Logo = req.Logo
	company.Address = req.Address
	company.BankName = req.BankName
	company.BankAccount = req.BankAccount
	company.BankHolder = req.BankHolder
	company.Email = req.Email
	company.Phone = req.Phone

	if req.ExpiredAt != nil {
		if *req.ExpiredAt != "" {
			t, err := time.Parse("2006-01-02", *req.ExpiredAt)
			if err != nil {
				return nil, fmt.Errorf("format tanggal expired_at salah")
			}
			company.ExpiredAt = &t
		} else {
			company.ExpiredAt = nil
		}
	}

	if err := s.db.WithContext(ctx).Save(&company).Error; err != nil {
		return nil, err
	}

	var adminEmail string
	s.db.WithContext(ctx).Model(&entity.User{}).Where("company_id = ?", company.ID).Order("id asc").Limit(1).Pluck("email", &adminEmail)

	res := toCompanyResponse(company, adminEmail)
	return &res, nil
}

func (s *companyService) Verify(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var company entity.Company
		if err := tx.First(&company, id).Error; err != nil {
			return err
		}

		if company.Status != "pending" {
			return fmt.Errorf("hanya perusahaan dengan status pending yang dapat diverifikasi")
		}

		company.Status = "verified"

		// Recalculate expired_at to start from the moment of approval (verification)
		var val string
		tx.Table("settings").Where("company_id = ? AND key = ?", 1, "default_active_days").Pluck("value", &val)
		days := 4 // Default fallback
		if val != "" {
			if parsedDays, err := strconv.Atoi(val); err == nil {
				days = parsedDays
			}
		}
		expiredAt := time.Now().AddDate(0, 0, days)
		company.ExpiredAt = &expiredAt

		if err := tx.Save(&company).Error; err != nil {
			return err
		}

		// Bootstrap defaults for this company
		if err := database.SeedCompanyDefaults(tx, company.ID); err != nil {
			return fmt.Errorf("gagal inisialisasi template data perusahaan: %w", err)
		}

		// Find registrant user (first user with RoleID is null in this company)
		var user entity.User
		if err := tx.Where("company_id = ? AND role_id IS NULL", company.ID).First(&user).Error; err == nil {
			var adminRole entity.Role
			if err := tx.Where("company_id = ? AND nama_role = ?", company.ID, "Admin").First(&adminRole).Error; err == nil {
				user.RoleID = &adminRole.ID
				if err := tx.Save(&user).Error; err != nil {
					return fmt.Errorf("gagal menetapkan role Admin ke registran: %w", err)
				}
			}
		}

		return nil
	})
}

func (s *companyService) Reject(ctx context.Context, id uint) error {
	var company entity.Company
	if err := s.db.WithContext(ctx).First(&company, id).Error; err != nil {
		return err
	}

	if company.Status != "pending" {
		return fmt.Errorf("hanya perusahaan dengan status pending yang dapat ditolak")
	}

	company.Status = "rejected"
	return s.db.WithContext(ctx).Save(&company).Error
}

func toCompanyResponse(c entity.Company, adminEmail string) dto.CompanyResponse {
	return dto.CompanyResponse{
		ID:           c.ID,
		Name:         c.Name,
		DirectorName: c.DirectorName,
		CeoNik:       c.CeoNik,
		Nib:          c.Nib,
		Logo:         c.Logo,
		Address:      c.Address,
		BankName:     c.BankName,
		BankAccount:  c.BankAccount,
		BankHolder:   c.BankHolder,
		Email:        c.Email,
		Phone:        c.Phone,
		Status:       c.Status,
		AdminEmail:   adminEmail,
		ExpiredAt:    c.ExpiredAt,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
