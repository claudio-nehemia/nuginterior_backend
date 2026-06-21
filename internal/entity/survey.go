package entity

import (
	"time"

	"gorm.io/datatypes"
)

// Survey represents the surveys table.
type Survey struct {
	ID                    uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID               uint           `gorm:"not null" json:"order_id"`
	TanggalSurvey         *time.Time     `json:"tanggal_survey,omitempty"`
	Lokasi                string         `gorm:"size:255" json:"lokasi"`
	Catatan               string         `gorm:"type:text" json:"catatan"`
	Status                string         `gorm:"size:50;not null;default:'pending'" json:"status"`
	SurveyorID            *uint          `json:"surveyor_id,omitempty"`
	ResponseBy            *string        `gorm:"size:255" json:"response_by,omitempty"`
	ResponseTime          *time.Time     `json:"response_time,omitempty"`
	MarketingResponseBy   *string        `gorm:"size:255" json:"marketing_response_by,omitempty"`
	MarketingResponseTime *time.Time     `json:"marketing_response_time,omitempty"`
	LayoutFiles           datatypes.JSON `gorm:"type:jsonb" json:"layout_files"`
	FotoLokasi            datatypes.JSON `gorm:"type:jsonb" json:"foto_lokasi"`
	MoMFile               string         `gorm:"size:500" json:"mom_file"`
	MomFiles              datatypes.JSON `gorm:"type:jsonb" json:"mom_files"`
	TanggalSurveyUlang    *time.Time     `json:"tanggal_survey_ulang,omitempty"`
	SurveyUlangTeamIDs    datatypes.JSON `gorm:"type:jsonb" json:"survey_ulang_team_ids"`
	CatatanUlang          string         `gorm:"type:text" json:"catatan_ulang"`
	TemuanLapangan        datatypes.JSON `gorm:"type:jsonb" json:"temuan_lapangan"`
	FotoVideoUlang        datatypes.JSON `gorm:"type:jsonb" json:"foto_video_ulang"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`

	Order            *Order             `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Surveyor         *User              `gorm:"foreignKey:SurveyorID" json:"surveyor,omitempty"`
	SurveyPengukuran []SurveyPengukuran `gorm:"foreignKey:SurveyID" json:"pengukuran,omitempty"`
}

func (Survey) TableName() string {
	return "surveys"
}
