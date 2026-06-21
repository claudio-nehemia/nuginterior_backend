package dto

import "time"

// ── Jenis Pengukuran Request DTOs ──

type CreateJenisPengukuranRequest struct {
	NamaPengukuran string `json:"nama_pengukuran" binding:"required,min=2,max=255"`
}

type UpdateJenisPengukuranRequest struct {
	NamaPengukuran string `json:"nama_pengukuran" binding:"required,min=2,max=255"`
}

// ── Jenis Pengukuran Response DTOs ──

type JenisPengukuranResponse struct {
	ID             uint      `json:"id"`
	NamaPengukuran string    `json:"nama_pengukuran"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
