package dto

import "time"

type CreateRABRequest struct {
	InputItemID   uint                  `json:"input_item_id" binding:"required"`
	OrderID       uint                  `json:"order_id" binding:"required"`
	MarkupGeneral float64               `json:"markup_general"`
	Rooms         []CreateRABRoomRequest `json:"rooms" binding:"required,dive"`
}

type CreateRABRoomRequest struct {
	NamaRuangan     string                         `json:"nama_ruangan" binding:"required"`
	ProdukID        *uint                          `json:"produk_id"`
	Qty             int                            `json:"qty" binding:"required,gt=0"`
	Panjang         float64                        `json:"panjang" binding:"required,gt=0"`
	Lebar           float64                        `json:"lebar" binding:"required,gt=0"`
	Tinggi          float64                        `json:"tinggi" binding:"required,gt=0"`
	Markup          float64                        `json:"markup"`
	BahanBakus      []CreateRABRoomBahanBakuRequest `json:"bahan_bakus"`
	FinishingDalams []CreateRABRoomFinishingRequest `json:"finishing_dalams"`
	FinishingLuars  []CreateRABRoomFinishingRequest `json:"finishing_luars"`
	Aksesoris       []CreateRABRoomAksesorisRequest `json:"aksesoris"`
}

type CreateRABRoomBahanBakuRequest struct {
	BahanBakuID uint    `json:"bahan_baku_id" binding:"required"`
	Markup      float64 `json:"markup"`
}

type CreateRABRoomFinishingRequest struct {
	ItemID uint    `json:"item_id" binding:"required"`
	Markup float64 `json:"markup"`
}

type CreateRABRoomAksesorisRequest struct {
	ItemID uint    `json:"item_id" binding:"required"`
	Qty    int     `json:"qty" binding:"required,gt=0"`
	Markup float64 `json:"markup"`
}

type UpdateRABRequest struct {
	MarkupGeneral float64               `json:"markup_general"`
	Rooms         []CreateRABRoomRequest `json:"rooms" binding:"required,dive"`
}

// RAB Responses
type RABResponse struct {
	ID            uint                    `json:"id"`
	InputItemID   uint                    `json:"input_item_id"`
	OrderID       uint                    `json:"order_id"`
	MarkupGeneral float64                 `json:"markup_general"`
	GrandTotal    float64                 `json:"grand_total"`
	Status        string                  `json:"status"` // 'draft', 'submitted'
	SubmittedAt   *time.Time              `json:"submitted_at"`
	SubmittedBy   string                  `json:"submitted_by"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	Order         *InputItemOrderResponse `json:"order,omitempty"`
	Rooms         []RABRoomResponse       `json:"rooms"`
}

type RABRoomResponse struct {
	ID              uint                         `json:"id"`
	NamaRuangan     string                       `json:"nama_ruangan"`
	ProdukID        *uint                        `json:"produk_id"`
	NamaProduk      string                       `json:"nama_produk"`
	Qty             int                          `json:"qty"`
	Panjang         float64                      `json:"panjang"`
	Lebar           float64                      `json:"lebar"`
	Tinggi          float64                      `json:"tinggi"`
	Markup          float64                      `json:"markup"`
	HargaDasar      float64                      `json:"harga_dasar"`
	HargaSatuan     float64                      `json:"harga_satuan"`
	HargaTotal      float64                      `json:"harga_total"`
	BahanBakus      []RABRoomBahanBakuResponse   `json:"bahan_bakus"`
	FinishingDalams []RABRoomFinishingResponse   `json:"finishing_dalams"`
	FinishingLuars  []RABRoomFinishingResponse   `json:"finishing_luars"`
	Aksesoris       []RABRoomAksesorisResponse   `json:"aksesoris"`
}

type RABRoomBahanBakuResponse struct {
	ID          uint    `json:"id"`
	BahanBakuID uint    `json:"bahan_baku_id"`
	NamaBahan   string  `json:"nama_bahan"`
	HargaDasar  float64 `json:"harga_dasar"`
	HargaJasa   float64 `json:"harga_jasa"`
	Markup      float64 `json:"markup"`
}

type RABRoomFinishingResponse struct {
	ID     uint    `json:"id"`
	ItemID uint    `json:"item_id"`
	Nama   string  `json:"nama_item"`
	Harga  float64 `json:"harga"`
	Type   string  `json:"type"` // "dalam" or "luar"
	Markup float64 `json:"markup"`
}

type RABRoomAksesorisResponse struct {
	ID         uint    `json:"id"`
	ItemID     uint    `json:"item_id"`
	Nama       string  `json:"nama_item"`
	Qty        int     `json:"qty"`
	Harga      float64 `json:"harga"`
	Markup     float64 `json:"markup"`
	HargaTotal float64 `json:"harga_total"`
}
