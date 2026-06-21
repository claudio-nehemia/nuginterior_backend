package dto

import "time"

type UpdateInvoiceDeadlineRequest struct {
	Deadline string `json:"deadline" binding:"required"` // Date string (e.g., "YYYY-MM-DD")
}

type InvoiceResponse struct {
	ID           uint       `json:"id"`
	ContractID   uint       `json:"contract_id"`
	OrderID      uint       `json:"order_id"`
	Step         int        `json:"step"`
	Keterangan   string     `json:"keterangan"`
	Persentase   float64    `json:"persentase"`
	Amount       float64    `json:"amount"`
	Deadline     *time.Time `json:"deadline"`
	Status       string     `json:"status"` // 'belum_bayar', 'terbayar'
	PaymentProof string     `json:"payment_proof"`
	PaidAt       *time.Time `json:"paid_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ContractInvoiceListResponse struct {
	ContractID          uint                `json:"contract_id"`
	OrderID             uint                `json:"order_id"`
	NomorOrder          string              `json:"nomor_order"`
	NamaProject         string              `json:"nama_project"`
	NamaCustomer        string              `json:"nama_customer"`
	TerminID            *uint               `json:"termin_id"`
	Termin              *TerminResponse     `json:"termin,omitempty"`
	StatusPembayaran    string              `json:"status_pembayaran"` // e.g. "Belum Bayar" or "DP (25%) Terbayar", dsb.
	InvoiceResponseBy   string              `json:"invoice_response_by"`
	InvoiceResponseTime *time.Time          `json:"invoice_response_time"`
	Invoices            []InvoiceResponse   `json:"invoices"`
}
