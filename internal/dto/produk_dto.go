package dto

import "time"

// ── Produk Request DTOs ──

type BahanBakuInput struct {
	BahanBakuID uint    `json:"bahan_baku_id" binding:"required"`
	HargaDasar  float64 `json:"harga_dasar" binding:"min=0"`
	HargaJasa   float64 `json:"harga_jasa" binding:"min=0"`
}

type CreateProdukRequest struct {
	NamaProduk string           `json:"nama_produk" binding:"required,min=2,max=255"`
	Images     []string         `json:"images" binding:"omitempty"`
	BahanBaku  []BahanBakuInput `json:"bahan_baku" binding:"omitempty"`
}

type UpdateProdukRequest struct {
	NamaProduk string           `json:"nama_produk" binding:"required,min=2,max=255"`
	Images     []string         `json:"images" binding:"omitempty"`
	BahanBaku  []BahanBakuInput `json:"bahan_baku" binding:"omitempty"`
}

// ── Produk Response DTOs ──

type ProdukImageResponse struct {
	ID    uint   `json:"id"`
	Image string `json:"image"`
}

type ProdukBahanBakuResponse struct {
	ID          uint    `json:"id"`
	BahanBakuID uint    `json:"bahan_baku_id"`
	NamaBahan   string  `json:"nama_bahan_baku"`
	HargaDasar  float64 `json:"harga_dasar"`
	HargaJasa   float64 `json:"harga_jasa"`
}

type ProdukResponse struct {
	ID         uint                      `json:"id"`
	NamaProduk string                    `json:"nama_produk"`
	Harga      float64                   `json:"harga"`
	HargaJasa  float64                   `json:"harga_jasa"`
	Images     []ProdukImageResponse     `json:"images,omitempty"`
	BahanBakus []ProdukBahanBakuResponse `json:"bahan_bakus,omitempty"`
	CreatedAt  time.Time                 `json:"created_at"`
	UpdatedAt  time.Time                 `json:"updated_at"`
}
