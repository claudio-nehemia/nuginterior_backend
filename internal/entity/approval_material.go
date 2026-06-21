package entity

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ApprovalMaterial struct {
	ID                    uint                   `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID               uint                   `gorm:"not null;uniqueIndex" json:"order_id"`
	Status                string                 `gorm:"size:50;not null;default:'pending'" json:"status"` // pending, completed
	ResponseBy            string                 `gorm:"size:255" json:"response_by"`
	ResponseTime          *time.Time             `json:"response_time"`
	MarketingResponseBy   string                 `gorm:"size:255" json:"marketing_response_by"`
	MarketingResponseTime *time.Time             `json:"marketing_response_time"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	DeletedAt             gorm.DeletedAt         `gorm:"index" json:"-"`

	// Relations
	Order *Order                 `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Items []ApprovalMaterialItem `gorm:"foreignKey:ApprovalMaterialID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (ApprovalMaterial) TableName() string {
	return "approval_materials"
}

type ApprovalMaterialItem struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ApprovalMaterialID uint           `gorm:"not null;index" json:"approval_material_id"`
	Category           string         `gorm:"size:50;not null" json:"category"` // "bahan_baku", "finishing_dalam", "finishing_luar", "aksesoris"
	SourceID           uint           `gorm:"not null" json:"source_id"`       // ID of BahanBaku or Item
	ItemName           string         `gorm:"size:255;not null" json:"item_name"`
	Area               string         `gorm:"type:text" json:"area"`
	Foto               string         `gorm:"size:255" json:"foto"`
	KodeMaterial       datatypes.JSON `gorm:"type:jsonb" json:"kode_material"` // JSON array of strings: e.g. ["TH 5027 NT", "TH 003 AA"]
	BrandSpek          datatypes.JSON `gorm:"type:jsonb" json:"brand_spek"`    // JSON array of strings: e.g. ["TACO", "HPL Taco"]
	Notes              string         `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func (ApprovalMaterialItem) TableName() string {
	return "approval_material_items"
}
