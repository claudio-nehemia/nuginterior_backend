package dto

import "time"

// ── Item Request DTOs ──

type CreateItemRequest struct {
	NamaItem  string  `json:"nama_item" binding:"required,min=2,max=255"`
	JenisItem string  `json:"jenis_item" binding:"required,oneof=finishing_dalam finishing_luar aksesoris"`
	Harga     float64 `json:"harga" binding:"min=0"`
}

type UpdateItemRequest struct {
	NamaItem  string  `json:"nama_item" binding:"required,min=2,max=255"`
	JenisItem string  `json:"jenis_item" binding:"required,oneof=finishing_dalam finishing_luar aksesoris"`
	Harga     float64 `json:"harga" binding:"min=0"`
}

// ── Item Response DTOs ──

type ItemResponse struct {
	ID        uint      `json:"id"`
	NamaItem  string    `json:"nama_item"`
	JenisItem string    `json:"jenis_item"`
	Harga     float64   `json:"harga"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
