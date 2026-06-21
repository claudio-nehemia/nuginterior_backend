package dto

import "time"

type CreateContractRequest struct {
	RABID       uint   `json:"rab_id" binding:"required"`
	TerminID    uint   `json:"termin_id" binding:"required"`
	LamaKontrak string `json:"lama_kontrak" binding:"required"`
}

type ContractResponse struct {
	ID                 uint                `json:"id"`
	RABID              uint                `json:"rab_id"`
	OrderID            uint                `json:"order_id"`
	TerminID           uint                `json:"termin_id"`
	LamaKontrak        string              `json:"lama_kontrak"`
	Status             string              `json:"status"` // 'draft', 'deal'
	SignedContractFile string              `json:"signed_contract_file"`
	ResponseBy         string              `json:"response_by"`
	ResponseTime       *time.Time          `json:"response_time"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	Order              *OrderBriefResponse `json:"order,omitempty"`
	Termin             *TerminResponse     `json:"termin,omitempty"`
	RAB                *RABResponse        `json:"rab,omitempty"`
}

type RABContractResponse struct {
	RABID        uint              `json:"rab_id"`
	OrderID      uint              `json:"order_id"`
	NomorOrder   string            `json:"nomor_order"`
	NamaProject  string            `json:"nama_project"`
	NamaCustomer string            `json:"nama_customer"`
	GrandTotal   float64           `json:"grand_total"` // Nilai kontrak
	Status       string            `json:"status"`      // 'belum_dibuat', 'draft', 'deal'
	ContractID   *uint             `json:"contract_id"`
	Contract     *ContractResponse `json:"contract,omitempty"`
}
