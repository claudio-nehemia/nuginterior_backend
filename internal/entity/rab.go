package entity

import (
	"time"

	"gorm.io/gorm"
)

// RAB represents the parent cost planning record linked to InputItem and Order.
type RAB struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	InputItemID   uint           `gorm:"not null;uniqueIndex" json:"input_item_id"`
	OrderID       uint           `gorm:"not null;index" json:"order_id"`
	MarkupGeneral float64        `gorm:"type:decimal(5,2);default:0" json:"markup_general"`
	GrandTotal    float64        `gorm:"type:decimal(18,2);default:0" json:"grand_total"`
	Status        string         `gorm:"type:varchar(50);default:'draft'" json:"status"` // 'draft', 'submitted'
	SubmittedAt   *time.Time     `json:"submitted_at"`
	SubmittedBy   string         `gorm:"type:varchar(100)" json:"submitted_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	InputItem *InputItem `gorm:"foreignKey:InputItemID" json:"input_item,omitempty"`
	Order     *Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Rooms     []RABRoom  `gorm:"foreignKey:RABID;constraint:OnDelete:CASCADE" json:"rooms,omitempty"`
	Contract  *Contract  `gorm:"foreignKey:RABID" json:"contract,omitempty"`
}

func (RAB) TableName() string {
	return "rabs"
}

// RABRoom represents a room in the RAB planning.
type RABRoom struct {
	ID          uint                 `gorm:"primaryKey;autoIncrement" json:"id"`
	RABID       uint                 `gorm:"not null;index" json:"rab_id"`
	NamaRuangan string               `gorm:"size:255;not null" json:"nama_ruangan"`
	ProdukID    *uint                `gorm:"null" json:"produk_id"`
	Qty         int                  `gorm:"not null;default:1" json:"qty"`
	Panjang     float64              `gorm:"type:decimal(10,2);not null;default:0" json:"panjang"`
	Lebar       float64              `gorm:"type:decimal(10,2);not null;default:0" json:"lebar"`
	Tinggi      float64              `gorm:"type:decimal(10,2);not null;default:0" json:"tinggi"`
	Markup      float64              `gorm:"type:decimal(5,2);default:0" json:"markup"` // Markup per produk
	HargaDasar  float64              `gorm:"type:decimal(18,2);default:0" json:"harga_dasar"`
	HargaSatuan float64              `gorm:"type:decimal(18,2);default:0" json:"harga_satuan"`
	HargaTotal  float64              `gorm:"type:decimal(18,2);default:0" json:"harga_total"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`

	// Relationships
	Produk          *Produk              `gorm:"foreignKey:ProdukID" json:"produk,omitempty"`
	BahanBakus      []RABRoomBahanBaku   `gorm:"foreignKey:RABRoomID;constraint:OnDelete:CASCADE" json:"bahan_bakus,omitempty"`
	FinishingDalams []RABRoomFinishing   `gorm:"foreignKey:RABRoomID;constraint:OnDelete:CASCADE" json:"finishing_dalams,omitempty"`
	FinishingLuars  []RABRoomFinishing   `gorm:"foreignKey:RABRoomID;constraint:OnDelete:CASCADE" json:"finishing_luars,omitempty"`
	Aksesoris       []RABRoomAksesoris   `gorm:"foreignKey:RABRoomID;constraint:OnDelete:CASCADE" json:"aksesoris,omitempty"`
}

func (RABRoom) TableName() string {
	return "rab_rooms"
}

// RABRoomBahanBaku represents the price mapping of materials in a RAB room.
type RABRoomBahanBaku struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RABRoomID   uint      `gorm:"not null;index" json:"rab_room_id"`
	BahanBakuID uint      `gorm:"not null" json:"bahan_baku_id"`
	HargaDasar  float64   `gorm:"type:decimal(18,2);not null;default:0" json:"harga_dasar"`
	HargaJasa   float64   `gorm:"type:decimal(18,2);not null;default:0" json:"harga_jasa"`
	Markup      float64   `gorm:"type:decimal(5,2);default:0" json:"markup"`
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	BahanBaku *BahanBaku `gorm:"foreignKey:BahanBakuID" json:"bahan_baku,omitempty"`
}

func (RABRoomBahanBaku) TableName() string {
	return "rab_room_bahan_bakus"
}

// RABRoomFinishing represents the finishing choice and price in a RAB room.
type RABRoomFinishing struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RABRoomID uint      `gorm:"not null;index" json:"rab_room_id"`
	ItemID    uint      `gorm:"not null" json:"item_id"`
	Type      string    `gorm:"size:50;not null" json:"type"` // "dalam" or "luar"
	Harga     float64   `gorm:"type:decimal(18,2);not null;default:0" json:"harga"`
	Markup    float64   `gorm:"type:decimal(5,2);default:0" json:"markup"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Item *Item `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}

func (RABRoomFinishing) TableName() string {
	return "rab_room_finishings"
}

// RABRoomAksesoris represents accessory items with dynamic markup and custom quantity in a RAB room.
type RABRoomAksesoris struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RABRoomID uint      `gorm:"not null;index" json:"rab_room_id"`
	ItemID    uint      `gorm:"not null" json:"item_id"`
	Qty       int       `gorm:"not null;default:1" json:"qty"`
	Harga     float64   `gorm:"type:decimal(18,2);not null;default:0" json:"harga"`
	Markup    float64   `gorm:"type:decimal(5,2);default:0" json:"markup"`
	HargaTotal float64  `gorm:"type:decimal(18,2);not null;default:0" json:"harga_total"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Item *Item `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}

func (RABRoomAksesoris) TableName() string {
	return "rab_room_aksesoris"
}
