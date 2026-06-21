package dto

import (
	"encoding/json"
	"time"
)

// Survey Request DTOs

type PengukuranInput struct {
	JenisPengukuranID *uint   `json:"jenis_pengukuran_id" binding:"omitempty"`
	NamaCustom        string  `json:"nama_custom" binding:"omitempty"`
	Checked           bool    `json:"checked"`
	Notes             string  `json:"notes" binding:"omitempty"`
	Panjang           float64 `json:"panjang" binding:"omitempty"`
	Lebar             float64 `json:"lebar" binding:"omitempty"`
	Tinggi            float64 `json:"tinggi" binding:"omitempty"`
	HasLebar          bool    `json:"has_lebar"`
	HasTinggi         bool    `json:"has_tinggi"`
}

type CreateSurveyRequest struct {
	OrderID            uint              `json:"order_id" binding:"required"`
	TanggalSurvey      string            `json:"tanggal_survey" binding:"omitempty"`
	Lokasi             string            `json:"lokasi" binding:"omitempty,max=255"`
	Catatan            string            `json:"catatan" binding:"omitempty"`
	Status             string            `json:"status" binding:"omitempty,oneof=pending dijadwalkan selesai batal"`
	SurveyorID         *uint             `json:"surveyor_id" binding:"omitempty"`
	LayoutFiles        json.RawMessage   `json:"layout_files" binding:"omitempty"`
	FotoLokasi         json.RawMessage   `json:"foto_lokasi" binding:"omitempty"`
	MoMFile            string            `json:"mom_file" binding:"omitempty"`
	MomFiles           json.RawMessage   `json:"mom_files" binding:"omitempty"`
	Pengukuran         []PengukuranInput `json:"pengukuran" binding:"omitempty"`
	TanggalSurveyUlang string            `json:"tanggal_survey_ulang" binding:"omitempty"`
	SurveyUlangTeamIDs json.RawMessage   `json:"survey_ulang_team_ids" binding:"omitempty"`
	CatatanUlang       string            `json:"catatan_ulang" binding:"omitempty"`
	TemuanLapangan     json.RawMessage   `json:"temuan_lapangan" binding:"omitempty"`
	FotoVideoUlang     json.RawMessage   `json:"foto_video_ulang" binding:"omitempty"`
}

type UpdateSurveyRequest struct {
	OrderID            uint              `json:"order_id" binding:"required"`
	TanggalSurvey      string            `json:"tanggal_survey" binding:"omitempty"`
	Lokasi             string            `json:"lokasi" binding:"omitempty,max=255"`
	Catatan            string            `json:"catatan" binding:"omitempty"`
	Status             string            `json:"status" binding:"omitempty,oneof=pending dijadwalkan selesai batal"`
	SurveyorID         *uint             `json:"surveyor_id" binding:"omitempty"`
	LayoutFiles        json.RawMessage   `json:"layout_files" binding:"omitempty"`
	FotoLokasi         json.RawMessage   `json:"foto_lokasi" binding:"omitempty"`
	MoMFile            string            `json:"mom_file" binding:"omitempty"`
	MomFiles           json.RawMessage   `json:"mom_files" binding:"omitempty"`
	Pengukuran         []PengukuranInput `json:"pengukuran" binding:"omitempty"`
	TanggalSurveyUlang string            `json:"tanggal_survey_ulang" binding:"omitempty"`
	SurveyUlangTeamIDs json.RawMessage   `json:"survey_ulang_team_ids" binding:"omitempty"`
	CatatanUlang       string            `json:"catatan_ulang" binding:"omitempty"`
	TemuanLapangan     json.RawMessage   `json:"temuan_lapangan" binding:"omitempty"`
	FotoVideoUlang     json.RawMessage   `json:"foto_video_ulang" binding:"omitempty"`
}

// Survey Response DTOs

type SurveyUserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
}

type PengukuranResponse struct {
	ID                uint    `json:"id"`
	JenisPengukuranID *uint   `json:"jenis_pengukuran_id"`
	NamaPengukuran    string  `json:"nama_pengukuran"`
	NamaCustom        string  `json:"nama_custom"`
	Checked           bool    `json:"checked"`
	Notes             string  `json:"notes"`
	Panjang           float64 `json:"panjang"`
	Lebar             float64 `json:"lebar"`
	Tinggi            float64 `json:"tinggi"`
	HasLebar          bool    `json:"has_lebar"`
	HasTinggi         bool    `json:"has_tinggi"`
}

type SurveyResponse struct {
	ID                    uint                 `json:"id"`
	OrderID               uint                 `json:"order_id"`
	TanggalSurvey         *time.Time           `json:"tanggal_survey,omitempty"`
	Lokasi                string               `json:"lokasi"`
	Catatan               string               `json:"catatan"`
	Status                string               `json:"status"`
	SurveyorID            *uint                `json:"surveyor_id,omitempty"`
	ResponseBy            *string              `json:"response_by,omitempty"`
	ResponseTime          *time.Time           `json:"response_time,omitempty"`
	MarketingResponseBy   *string              `json:"marketing_response_by,omitempty"`
	MarketingResponseTime *time.Time           `json:"marketing_response_time,omitempty"`
	LayoutFiles           json.RawMessage      `json:"layout_files"`
	FotoLokasi            json.RawMessage      `json:"foto_lokasi"`
	MoMFile               string               `json:"mom_file"`
	MomFiles              json.RawMessage      `json:"mom_files"`
	TanggalSurveyUlang    *time.Time           `json:"tanggal_survey_ulang,omitempty"`
	SurveyUlangTeamIDs    json.RawMessage      `json:"survey_ulang_team_ids"`
	CatatanUlang          string               `json:"catatan_ulang"`
	TemuanLapangan        json.RawMessage      `json:"temuan_lapangan"`
	FotoVideoUlang        json.RawMessage      `json:"foto_video_ulang"`
	IsContractDeal        bool                 `json:"is_contract_deal"`
	Surveyor              *SurveyUserResponse  `json:"surveyor,omitempty"`
	SurveyUlangTeam       []SurveyUserResponse `json:"survey_ulang_team,omitempty"`
	Order                 *OrderBriefResponse  `json:"order,omitempty"`
	Pengukuran            []PengukuranResponse `json:"pengukuran,omitempty"`
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
}
