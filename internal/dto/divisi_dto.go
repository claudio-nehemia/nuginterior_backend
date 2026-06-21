package dto

import "time"

// ── Divisi Request DTOs ──

type CreateDivisiRequest struct {
	NamaDivisi string `json:"nama_divisi" binding:"required,min=2,max=255"`
}

type UpdateDivisiRequest struct {
	NamaDivisi string `json:"nama_divisi" binding:"required,min=2,max=255"`
}

// ── Divisi Response DTOs ──

type DivisiResponse struct {
	ID         uint      `json:"id"`
	NamaDivisi string    `json:"nama_divisi"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
