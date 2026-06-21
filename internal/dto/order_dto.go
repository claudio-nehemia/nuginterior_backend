package dto

import (
	"encoding/json"
	"time"
)

// Order Request DTOs

type CreateOrderRequest struct {
	NamaProject            string `json:"nama_project" binding:"required,min=2,max=255"`
	JenisInterior          string `json:"jenis_interior" binding:"omitempty,oneof=residential apartment office retail hospitality custom"`
	NamaCustomer           string `json:"nama_customer" binding:"required,min=2,max=255"`
	TeleponCustomer        string `json:"telepon_customer" binding:"omitempty,max=50"`
	EmailCustomer          string `json:"email_customer" binding:"omitempty,max=255"`
	NamaPerusahaan         string `json:"nama_perusahaan" binding:"omitempty,max=255"`
	CustomerAdditionalInfo string `json:"customer_additional_info" binding:"omitempty"`
	NomorUnit              string `json:"nomor_unit" binding:"omitempty,max=255"`
	Alamat                 string `json:"alamat" binding:"omitempty"`
	Catatan                string `json:"catatan" binding:"omitempty"`
	TanggalMasukCustomer   string `json:"tanggal_masuk_customer" binding:"omitempty"`
	PriorityLevel          string `json:"priority_level" binding:"omitempty,oneof=low medium high"`
	TanggalSurvey          string `json:"tanggal_survey" binding:"omitempty"`
}

type UpdateOrderRequest struct {
	NamaProject            string  `json:"nama_project" binding:"required,min=2,max=255"`
	JenisInterior          string  `json:"jenis_interior" binding:"omitempty,oneof=residential apartment office retail hospitality custom"`
	NamaCustomer           string  `json:"nama_customer" binding:"required,min=2,max=255"`
	TeleponCustomer        string  `json:"telepon_customer" binding:"omitempty,max=50"`
	EmailCustomer          string  `json:"email_customer" binding:"omitempty,max=255"`
	NamaPerusahaan         string  `json:"nama_perusahaan" binding:"omitempty,max=255"`
	CustomerAdditionalInfo string  `json:"customer_additional_info" binding:"omitempty"`
	NomorUnit              string  `json:"nomor_unit" binding:"omitempty,max=255"`
	Alamat                 string  `json:"alamat" binding:"omitempty"`
	Catatan                string  `json:"catatan" binding:"omitempty"`
	TanggalMasukCustomer   string  `json:"tanggal_masuk_customer" binding:"omitempty"`
	ProjectStatus          string  `json:"project_status" binding:"omitempty,oneof=pending in_progress deal cancel"`
	PriorityLevel          string  `json:"priority_level" binding:"omitempty,oneof=low medium high"`
	TahapanProyek          string  `json:"tahapan_proyek" binding:"omitempty"`
	PaymentStatus          string  `json:"payment_status" binding:"omitempty,oneof=not_start cm_fee dp termin lunas"`
	TerminID               *uint   `json:"termin_id" binding:"omitempty"`
	PicID                  *uint   `json:"pic_id" binding:"omitempty"`
	HargaKontrak           float64 `json:"harga_kontrak" binding:"omitempty"`
	TanggalKontrak         string  `json:"tanggal_kontrak" binding:"omitempty"`
	NomorKontrak           string  `json:"nomor_kontrak" binding:"omitempty,max=100"`
	TanggalMulai           string  `json:"tanggal_mulai" binding:"omitempty"`
	TanggalSelesai         string  `json:"tanggal_selesai" binding:"omitempty"`
	TanggalSurvey          string  `json:"tanggal_survey" binding:"omitempty"`
}

type SyncTeamRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
}

// Order Response DTOs

type OrderBriefResponse struct {
	ID                   uint                `json:"id"`
	NomorOrder           string              `json:"nomor_order"`
	NamaProject          string              `json:"nama_project"`
	NamaCustomer         string              `json:"nama_customer"`
	NamaPerusahaan       string              `json:"nama_perusahaan"`
	JenisInterior        string              `json:"jenis_interior"`
	Alamat               string              `json:"alamat"`
	TahapanProyek        string              `json:"tahapan_proyek"`
	PaymentStatus        string              `json:"payment_status"`
	TanggalMasukCustomer *time.Time          `json:"tanggal_masuk_customer,omitempty"`
	Teams                []OrderTeamResponse `json:"teams,omitempty"`
	PicName              string              `json:"pic_name,omitempty"`
	ProductCount         int                 `json:"product_count,omitempty"`
	LamaKontrak          string              `json:"lama_kontrak,omitempty"`
}

type OrderTeamResponse struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type OrderResponse struct {
	ID                     uint                `json:"id"`
	NomorOrder             string              `json:"nomor_order"`
	NamaProject            string              `json:"nama_project"`
	JenisInterior          string              `json:"jenis_interior"`
	NamaCustomer           string              `json:"nama_customer"`
	TeleponCustomer        string              `json:"telepon_customer"`
	EmailCustomer          string              `json:"email_customer"`
	NamaPerusahaan         string              `json:"nama_perusahaan"`
	CustomerAdditionalInfo string              `json:"customer_additional_info"`
	NomorUnit              string              `json:"nomor_unit"`
	Alamat                 string              `json:"alamat"`
	Catatan                string              `json:"catatan"`
	TanggalMasukCustomer   *time.Time          `json:"tanggal_masuk_customer,omitempty"`
	ProjectStatus          string              `json:"project_status"`
	PriorityLevel          string              `json:"priority_level"`
	TahapanProyek          string              `json:"tahapan_proyek"`
	PaymentStatus          string              `json:"payment_status"`
	TerminID               *uint               `json:"termin_id,omitempty"`
	HargaKontrak           string              `json:"harga_kontrak"`
	TanggalKontrak         *time.Time          `json:"tanggal_kontrak,omitempty"`
	NomorKontrak           string              `json:"nomor_kontrak"`
	TanggalMulai           *time.Time          `json:"tanggal_mulai,omitempty"`
	TanggalSelesai         *time.Time          `json:"tanggal_selesai,omitempty"`
	PicID                  *uint               `json:"pic_id,omitempty"`
	MomFile                string              `json:"mom_file"`
	MomFiles               json.RawMessage     `json:"mom_files"`
	TanggalSurvey          string              `json:"tanggal_survey"`
	SurveyResponseBy       string              `json:"survey_response_by"`
	SurveyResponseTime     string              `json:"survey_response_time"`
	PmSurveyResponseBy     string              `json:"pm_survey_response_by"`
	PmSurveyResponseTime   string              `json:"pm_survey_response_time"`
	Teams                  []OrderTeamResponse `json:"teams,omitempty"`
	Surveys                []SurveyResponse    `json:"surveys,omitempty"`
	Moodboards             []MoodboardResponse `json:"moodboards,omitempty"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
}
