package entity

import (
	"time"

	"gorm.io/gorm"
)

// Setting represents the settings table.
type Setting struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID   uint      `gorm:"not null;default:1;uniqueIndex:idx_company_key" json:"company_id"`
	Key         string    `gorm:"size:255;not null;uniqueIndex:idx_company_key" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Setting) TableName() string {
	return "settings"
}

// CompanyProfile represents the company profile configuration fields.
type CompanyProfile struct {
	Name        string
	Director    string
	Logo        string
	Address     string
	BankName    string
	BankAccount string
	BankHolder  string
	Email       string
	Phone       string
}

// GetCompanyProfile fetches all settings with key starting with "company_" from DB
// and constructs a CompanyProfile object with standard fallbacks.
func GetCompanyProfile(db *gorm.DB, companyID uint) CompanyProfile {
	var settings []Setting
	_ = db.Where("company_id = ? AND key LIKE ?", companyID, "company_%").Find(&settings)

	profile := CompanyProfile{
		Name:        "PT. INTERIORPRO INDONESIA",
		Director:    "Super Admin",
		Logo:        "",
		Address:     "Alamat Perusahaan",
		BankName:    "BANK CENTRAL ASIA (BCA)",
		BankAccount: "000-000-0000",
		BankHolder:  "PT. INTERIORPRO INDONESIA",
		Email:       "finance@interiorpro.com",
		Phone:       "+62 812-0000-0000",
	}

	for _, s := range settings {
		switch s.Key {
		case "company_name":
			profile.Name = s.Value
		case "company_director":
			profile.Director = s.Value
		case "company_logo":
			profile.Logo = s.Value
		case "company_address":
			profile.Address = s.Value
		case "company_bank_name":
			profile.BankName = s.Value
		case "company_bank_account":
			profile.BankAccount = s.Value
		case "company_bank_holder":
			profile.BankHolder = s.Value
		case "company_email":
			profile.Email = s.Value
		case "company_phone":
			profile.Phone = s.Value
		}
	}
	return profile
}
