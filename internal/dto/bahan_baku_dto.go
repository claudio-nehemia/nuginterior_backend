package dto

import "time"

// ── Bahan Baku Request DTOs ──

type CreateBahanBakuRequest struct {
	NamaBahanBaku string `json:"nama_bahan_baku" binding:"required,min=2,max=255"`
}

type UpdateBahanBakuRequest struct {
	NamaBahanBaku string `json:"nama_bahan_baku" binding:"required,min=2,max=255"`
}

// ── Bahan Baku Response DTOs ──

type BahanBakuResponse struct {
	ID            uint      `json:"id"`
	NamaBahanBaku string    `json:"nama_bahan_baku"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
