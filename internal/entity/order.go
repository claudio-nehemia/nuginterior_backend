package entity

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// Order represents the orders table.
type Order struct {
	ID                     uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID              uint            `gorm:"not null;default:1" json:"company_id"`
	NomorOrder             string          `gorm:"size:50;uniqueIndex" json:"nomor_order"`
	NamaProject            string          `gorm:"size:255;not null" json:"nama_project"`
	JenisInterior          string          `gorm:"size:50;not null;default:'residential'" json:"jenis_interior"`
	NamaCustomer           string          `gorm:"size:255;not null" json:"nama_customer"`
	TeleponCustomer        string          `gorm:"size:50" json:"telepon_customer"`
	EmailCustomer          string          `gorm:"size:255" json:"email_customer"`
	NamaPerusahaan         string          `gorm:"size:255" json:"nama_perusahaan"`
	CustomerAdditionalInfo string          `gorm:"type:text" json:"customer_additional_info"`
	NomorUnit              string          `gorm:"size:255" json:"nomor_unit"`
	Alamat                 string          `gorm:"type:text" json:"alamat"`
	Catatan                string          `gorm:"type:text" json:"catatan"`
	TanggalMasukCustomer   *time.Time      `json:"tanggal_masuk_customer,omitempty"`
	ProjectStatus          string          `gorm:"size:50;default:'pending'" json:"project_status"`
	PriorityLevel          string          `gorm:"size:20;default:'medium'" json:"priority_level"`
	TahapanProyek          string          `gorm:"size:50;default:'not_start'" json:"tahapan_proyek"`
	PaymentStatus          string          `gorm:"size:50;default:'not_start'" json:"payment_status"`
	TerminID               *uint           `json:"termin_id,omitempty"`
	HargaKontrak           decimal.Decimal `gorm:"type:decimal(18,2);default:0" json:"harga_kontrak"`
	TanggalKontrak         *time.Time      `json:"tanggal_kontrak,omitempty"`
	NomorKontrak           string          `gorm:"size:100" json:"nomor_kontrak"`
	TanggalMulai           *time.Time      `json:"tanggal_mulai,omitempty"`
	TanggalSelesai         *time.Time      `json:"tanggal_selesai,omitempty"`
	PicID                  *uint           `json:"pic_id,omitempty"`
	MomFile                string          `gorm:"size:500" json:"mom_file"`
	MomFiles               datatypes.JSON  `gorm:"type:jsonb" json:"mom_files"`
	TanggalSurvey          string          `gorm:"size:255" json:"tanggal_survey"`
	SurveyResponseBy       string          `gorm:"size:255" json:"survey_response_by"`
	SurveyResponseTime     string          `gorm:"size:255" json:"survey_response_time"`
	PmSurveyResponseBy     string          `gorm:"size:255" json:"pm_survey_response_by"`
	PmSurveyResponseTime   string          `gorm:"size:255" json:"pm_survey_response_time"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`

	// Relations
	Company    *Company     `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Termin     *Termin      `gorm:"foreignKey:TerminID" json:"termin,omitempty"`
	PIC        *User        `gorm:"foreignKey:PicID" json:"pic,omitempty"`
	Surveys    []Survey     `gorm:"foreignKey:OrderID" json:"surveys,omitempty"`
	Moodboards []Moodboard  `gorm:"foreignKey:OrderID" json:"moodboards,omitempty"`
	Teams      []OrderTeam  `gorm:"foreignKey:OrderID" json:"teams,omitempty"`
	Contracts  []Contract   `gorm:"foreignKey:OrderID" json:"contracts,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}
