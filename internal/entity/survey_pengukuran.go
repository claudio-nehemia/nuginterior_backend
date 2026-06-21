package entity

import "time"

// SurveyPengukuran represents the survey_pengukuran pivot table.
type SurveyPengukuran struct {
	ID                uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	SurveyID          uint             `gorm:"not null" json:"survey_id"`
	JenisPengukuranID *uint            `json:"jenis_pengukuran_id"`
	NamaCustom        string           `gorm:"size:255" json:"nama_custom"`
	Checked           bool             `gorm:"not null;default:false" json:"checked"`
	Notes             string           `gorm:"type:text" json:"notes"`
	Panjang           float64          `gorm:"type:decimal(10,2);default:0" json:"panjang"`
	Lebar             float64          `gorm:"type:decimal(10,2);default:0" json:"lebar"`
	Tinggi            float64          `gorm:"type:decimal(10,2);default:0" json:"tinggi"`
	HasLebar          bool             `gorm:"default:false" json:"has_lebar"`
	HasTinggi         bool             `gorm:"default:false" json:"has_tinggi"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`

	JenisPengukuran   *JenisPengukuran `gorm:"foreignKey:JenisPengukuranID" json:"jenis_pengukuran,omitempty"`
}

func (SurveyPengukuran) TableName() string {
	return "survey_pengukuran"
}
