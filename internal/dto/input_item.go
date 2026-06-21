package dto

import "time"

type CreateInputItemRequest struct {
	DesainFinalID uint                `json:"desain_final_id" binding:"required"`
	OrderID       uint                `json:"order_id" binding:"required"`
	Status        string              `json:"status" binding:"required,oneof=draft approved"`
	Rooms         []CreateRoomRequest `json:"rooms" binding:"required,dive"`
}

type CreateRoomRequest struct {
	NamaRuangan     string                   `json:"nama_ruangan" binding:"required"`
	ProdukID        *uint                    `json:"produk_id"`
	Qty             int                      `json:"qty"`
	Panjang         float64                  `json:"panjang"`
	Lebar           float64                  `json:"lebar"`
	Tinggi          float64                  `json:"tinggi"`
	BahanBakus      []uint                   `json:"bahan_bakus"` // Array of BahanBakuID
	FinishingDalams []CreateFinishingRequest `json:"finishing_dalams"`
	FinishingLuars  []CreateFinishingRequest `json:"finishing_luars"`
	Aksesoris       []CreateAksesorisRequest `json:"aksesoris"`
}

type CreateFinishingRequest struct {
	ItemID uint   `json:"item_id" binding:"required"`
	Notes  string `json:"notes"`
}

type CreateAksesorisRequest struct {
	ItemID uint   `json:"item_id" binding:"required"`
	Qty    int    `json:"qty" binding:"required,gt=0"`
	Notes  string `json:"notes"`
}

type UpdateInputItemRequest struct {
	Status string              `json:"status" binding:"required,oneof=draft approved"`
	Rooms  []CreateRoomRequest `json:"rooms" binding:"required,dive"`
}

type InputItemOrderResponse struct {
	ID             uint   `json:"id"`
	NomorOrder     string `json:"nomor_order"`
	NamaProject    string `json:"nama_project"`
	NamaCustomer   string `json:"nama_customer"`
	NamaPerusahaan string `json:"nama_perusahaan,omitempty"`
	JenisInterior  string `json:"jenis_interior"`
}

// Responses
type InputItemResponse struct {
	ID                    uint                    `json:"id"`
	DesainFinalID         uint                    `json:"desain_final_id"`
	OrderID               uint                    `json:"order_id"`
	Status                string                  `json:"status"`
	ResponseTime          *time.Time              `json:"response_time"`
	ResponseBy            string                  `json:"response_by"`
	MarketingResponseTime *time.Time              `json:"marketing_response_time"`
	MarketingResponseBy   string                  `json:"marketing_response_by"`
	CreatedAt             time.Time               `json:"created_at"`
	UpdatedAt             time.Time               `json:"updated_at"`
	Order                 *InputItemOrderResponse `json:"order,omitempty"`
	Rooms                 []RoomResponse          `json:"rooms"`
}

type RoomResponse struct {
	ID              uint                    `json:"id"`
	NamaRuangan     string                  `json:"nama_ruangan"`
	ProdukID        *uint                   `json:"produk_id"`
	NamaProduk      string                  `json:"nama_produk"`
	Qty             int                     `json:"qty"`
	Panjang         float64                 `json:"panjang"`
	Lebar           float64                 `json:"lebar"`
	Tinggi          float64                 `json:"tinggi"`
	BahanBakus      []RoomBahanBakuResponse `json:"bahan_bakus"`
	FinishingDalams []RoomFinishingResponse `json:"finishing_dalams"`
	FinishingLuars  []RoomFinishingResponse `json:"finishing_luars"`
	Aksesoris       []RoomAksesorisResponse `json:"aksesoris"`
}

type RoomBahanBakuResponse struct {
	ID          uint   `json:"id"`
	BahanBakuID uint   `json:"bahan_baku_id"`
	NamaBahan   string `json:"nama_bahan"`
}

type RoomFinishingResponse struct {
	ID     uint   `json:"id"`
	ItemID uint   `json:"item_id"`
	Nama   string `json:"nama_item"`
	Notes  string `json:"notes"`
}

type RoomAksesorisResponse struct {
	ID     uint   `json:"id"`
	ItemID uint   `json:"item_id"`
	Nama   string `json:"nama_item"`
	Qty    int    `json:"qty"`
	Notes  string `json:"notes"`
}
