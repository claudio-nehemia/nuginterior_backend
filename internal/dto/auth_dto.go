package dto

import "time"

// ── Auth Request DTOs ──

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	RoleID   *uint  `json:"role_id" binding:"omitempty"`
}

type RegisterCompanyRequest struct {
	UserName     string `json:"user_name" binding:"required,min=2,max=255"`
	UserEmail    string `json:"user_email" binding:"required,email"`
	UserPassword string `json:"user_password" binding:"required,min=8"`

	CompanyName    string `json:"company_name" binding:"required,min=2,max=255"`
	DirectorName   string `json:"director_name" binding:"required,min=2,max=255"`
	CeoNik         string `json:"ceo_nik" binding:"required,min=3,max=50"`
	Nib            string `json:"nib" binding:"omitempty"`
	CompanyEmail   string `json:"company_email" binding:"required,email"`
	CompanyPhone   string `json:"company_phone" binding:"required"`
	CompanyAddress string `json:"company_address" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ── Auth Response DTOs ──

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type UserProfileResponse struct {
	ID              uint              `json:"id"`
	CompanyID       uint              `json:"company_id"`
	Company         *CompanySimple    `json:"company,omitempty"`
	Name            string            `json:"name"`
	Email           string            `json:"email"`
	EmailVerifiedAt *time.Time        `json:"email_verified_at,omitempty"`
	RoleID          *uint             `json:"role_id,omitempty"`
	Role            *RoleSimple       `json:"role,omitempty"`
	Permissions     []string          `json:"permissions"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type RoleSimple struct {
	ID       uint   `json:"id"`
	NamaRole string `json:"nama_role"`
	DivisiID uint   `json:"divisi_id"`
}

type CompanySimple struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Status string `json:"status"`
}
