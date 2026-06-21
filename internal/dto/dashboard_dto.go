package dto

import "time"

type RecentOrderResponse struct {
	ID            uint      `json:"id"`
	NomorOrder    string    `json:"nomor_order"`
	NamaProject   string    `json:"nama_project"`
	NamaCustomer  string    `json:"nama_customer"`
	ProjectStatus string    `json:"project_status"`
	TahapanProyek string    `json:"tahapan_proyek"`
	HargaKontrak  float64   `json:"harga_kontrak"`
	CreatedAt     time.Time `json:"created_at"`
}

type DashboardStatsResponse struct {
	TotalOrders        int64                 `json:"total_orders"`
	ActiveOrders       int64                 `json:"active_orders"`
	CompletedProjects  int64                 `json:"completed_projects"`
	SuccessRate        float64               `json:"success_rate"`
	TotalContractsDeal int64                 `json:"total_contracts_deal"`
	LunasCount         int64                 `json:"lunas_count"`
	LunasAmount        float64               `json:"lunas_amount"`
	BelumBayarCount    int64                 `json:"belum_bayar_count"`
	BelumBayarAmount   float64               `json:"belum_bayar_amount"`
	TotalOmset         float64               `json:"total_omset"`
	RecentOrders       []RecentOrderResponse `json:"recent_orders"`
}
