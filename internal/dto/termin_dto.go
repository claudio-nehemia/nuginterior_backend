package dto

import (
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
)

// ── Termin Request DTOs ──

type TahapanInput struct {
	Step       int     `json:"step" binding:"required"`
	Text       string  `json:"text" binding:"required"`
	Persentase float64 `json:"persentase" binding:"required,min=0,max=100"`
}

type CreateTerminRequest struct {
	KodeTipe  string         `json:"kode_tipe" binding:"required,min=1,max=255"`
	NamaTipe  string         `json:"nama_tipe" binding:"required,min=2,max=255"`
	Deskripsi string         `json:"deskripsi" binding:"omitempty"`
	Tahapan   []TahapanInput `json:"tahapan" binding:"required,min=1"`
}

type UpdateTerminRequest struct {
	KodeTipe  string         `json:"kode_tipe" binding:"required,min=1,max=255"`
	NamaTipe  string         `json:"nama_tipe" binding:"required,min=2,max=255"`
	Deskripsi string         `json:"deskripsi" binding:"omitempty"`
	Tahapan   []TahapanInput `json:"tahapan" binding:"required,min=1"`
}

// ── Termin Response DTOs ──

type TahapanResponse struct {
	Step       int     `json:"step"`
	Text       string  `json:"text"`
	Persentase float64 `json:"persentase"`
}

type TerminResponse struct {
	ID        uint              `json:"id"`
	KodeTipe  string            `json:"kode_tipe"`
	NamaTipe  string            `json:"nama_tipe"`
	Deskripsi string            `json:"deskripsi"`
	Tahapan   []TahapanResponse `json:"tahapan"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// ToTahapanEntity converts DTO inputs to entity Tahapans.
func ToTahapanEntity(inputs []TahapanInput) entity.Tahapans {
	result := make(entity.Tahapans, len(inputs))
	for i, input := range inputs {
		result[i] = entity.TahapanItem{
			Step:       input.Step,
			Text:       input.Text,
			Persentase: input.Persentase,
		}
	}
	return result
}
