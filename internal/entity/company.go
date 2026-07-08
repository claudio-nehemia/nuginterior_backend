package entity

import (
	"time"
)

// Company represents the companies table.
type Company struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	DirectorName string    `gorm:"size:255;not null" json:"director_name"`
	CeoNik       string    `gorm:"size:50;not null" json:"ceo_nik"`
	Nib          string    `gorm:"size:50" json:"nib"`
	Logo         string    `gorm:"size:500" json:"logo"`
	Address      string    `gorm:"type:text" json:"address"`
	BankName     string    `gorm:"size:255" json:"bank_name"`
	BankAccount  string    `gorm:"size:100" json:"bank_account"`
	BankHolder   string    `gorm:"size:255" json:"bank_holder"`
	Email        string    `gorm:"size:255" json:"email"`
	Phone        string    `gorm:"size:50" json:"phone"`
	Status       string    `gorm:"size:50;not null;default:'pending'" json:"status"` // pending, verified, rejected
	ExpiredAt    *time.Time `json:"expired_at" gorm:"default:null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Company) TableName() string {
	return "companies"
}
