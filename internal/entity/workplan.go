package entity

import (
	"time"

	"gorm.io/gorm"
)

type Workplan struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID      uint           `gorm:"not null;uniqueIndex" json:"order_id"`
	Status       string         `gorm:"size:50;not null;default:'draft'" json:"status"` // draft, submitted
	StartDate    *time.Time     `json:"start_date"`
	EndDate      *time.Time     `json:"end_date"`
	DurationDays int            `gorm:"not null;default:0" json:"duration_days"`
	ResponseBy      string         `gorm:"size:255" json:"response_by"`
	ResponseTime    *time.Time     `json:"response_time"`
	ExtensionStatus string         `gorm:"size:50;not null;default:'none'" json:"extension_status"` // none, pending, approved, rejected
	ExtensionNotes  string         `gorm:"type:text" json:"extension_notes"`
	ExtensionDays   int            `gorm:"not null;default:0" json:"extension_days"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Order  *Order          `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Stages []WorkplanStage `gorm:"foreignKey:WorkplanID;constraint:OnDelete:CASCADE" json:"stages,omitempty"`

	// BAST Fields
	BastPhoto       string     `gorm:"type:text" json:"bast_photo"`
	BastGeneratedAt *time.Time `json:"bast_generated_at"`
	BastGeneratedBy string     `gorm:"size:255" json:"bast_generated_by"`
}

func (Workplan) TableName() string {
	return "workplans"
}

type WorkplanStageMaster struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code       string    `gorm:"size:100;uniqueIndex;not null" json:"code"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	Percentage float64   `gorm:"type:decimal(5,2);not null;default:0" json:"percentage"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (WorkplanStageMaster) TableName() string {
	return "workplan_stage_masters"
}

type WorkplanStage struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkplanID      uint       `gorm:"not null;index" json:"workplan_id"`
	InputItemRoomID uint       `gorm:"not null;index" json:"input_item_room_id"`
	StageMasterID   uint       `gorm:"not null;index" json:"stage_master_id"`
	Percentage      float64    `gorm:"type:decimal(5,2);not null;default:0" json:"percentage"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	Notes           string     `gorm:"type:text" json:"notes"`
	Status          string     `gorm:"size:50;not null;default:'pending'" json:"status"` // pending, completed
	CompletedAt     *time.Time `json:"completed_at"`
	CompletedBy     string     `gorm:"size:255" json:"completed_by"`
	Photos          string     `gorm:"type:text" json:"photos"` // comma-separated photo URLs
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	Workplan        *Workplan            `gorm:"foreignKey:WorkplanID" json:"workplan,omitempty"`
	InputItemRoom   *InputItemRoom       `gorm:"foreignKey:InputItemRoomID" json:"input_item_room,omitempty"`
	StageMaster     *WorkplanStageMaster `gorm:"foreignKey:StageMasterID;constraint:OnDelete:CASCADE" json:"stage_master,omitempty"`
}

func (WorkplanStage) TableName() string {
	return "workplan_stages"
}

type WorkplanDefect struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkplanStageID uint       `gorm:"not null;index" json:"workplan_stage_id"`
	Description     string     `gorm:"type:text;not null" json:"description"`
	Photos          string     `gorm:"type:text" json:"photos"` // comma-separated defect photos
	Status          string     `gorm:"size:50;not null;default:'reported'" json:"status"` // reported | fix_submitted | accepted | rejected
	FixDescription  string     `gorm:"type:text" json:"fix_description"`
	FixPhotos       string     `gorm:"type:text" json:"fix_photos"` // comma-separated fix photos
	ReportedBy      string     `gorm:"size:255" json:"reported_by"`
	FixedBy         string     `gorm:"size:255" json:"fixed_by"`
	ReviewedBy      string     `gorm:"size:255" json:"reviewed_by"`
	ReviewNotes     string     `gorm:"type:text" json:"review_notes"`
	ReportedAt      *time.Time `json:"reported_at"`
	FixedAt         *time.Time `json:"fixed_at"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	WorkplanStage *WorkplanStage `gorm:"foreignKey:WorkplanStageID" json:"workplan_stage,omitempty"`
}

func (WorkplanDefect) TableName() string {
	return "workplan_defects"
}

