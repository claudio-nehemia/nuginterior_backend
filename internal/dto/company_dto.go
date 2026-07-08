package dto

import "time"

type CompanyUpdateRequest struct {
	Name         string  `json:"name" binding:"required,min=2,max=255"`
	DirectorName string  `json:"director_name" binding:"required,min=2,max=255"`
	CeoNik       string  `json:"ceo_nik" binding:"required,min=3,max=50"`
	Nib          string  `json:"nib" binding:"omitempty"`
	Logo         string  `json:"logo" binding:"omitempty"`
	Address      string  `json:"address" binding:"required"`
	BankName     string  `json:"bank_name" binding:"omitempty"`
	BankAccount  string  `json:"bank_account" binding:"omitempty"`
	BankHolder   string  `json:"bank_holder" binding:"omitempty"`
	Email        string  `json:"email" binding:"required,email"`
	Phone        string  `json:"phone" binding:"required"`
	ExpiredAt    *string `json:"expired_at" binding:"omitempty"`
}

type CompanyResponse struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	DirectorName string     `json:"director_name"`
	CeoNik       string     `json:"ceo_nik"`
	Nib          string     `json:"nib"`
	Logo         string     `json:"logo"`
	Address      string     `json:"address"`
	BankName     string     `json:"bank_name"`
	BankAccount  string     `json:"bank_account"`
	BankHolder   string     `json:"bank_holder"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Status       string     `json:"status"`
	AdminEmail   string     `json:"admin_email"`
	ExpiredAt    *time.Time `json:"expired_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
