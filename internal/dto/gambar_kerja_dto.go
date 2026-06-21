package dto

import "time"

// GambarKerja Request DTOs
type CreateGambarKerjaRequest struct {
	OrderID uint `json:"order_id" binding:"required"`
}

type ReviseWorkingDrawingFileRequest struct {
	FileID uint   `json:"file_id" binding:"required"`
	Notes  string `json:"notes" binding:"required"`
}

type ReviseGeneralRequest struct {
	Notes string `json:"notes" binding:"required"`
}

// GambarKerja Response DTOs
type GambarKerjaFileResponse struct {
	ID            uint      `json:"id"`
	GambarKerjaID uint      `json:"gambar_kerja_id"`
	FilePath      string    `json:"file_path"`
	OriginalName  string    `json:"original_name"`
	Status        string    `json:"status"`
	Revisi        string    `json:"revisi"`
	CreatedAt     time.Time `json:"created_at"`
}

type GambarKerjaResponse struct {
	ID                    uint                      `json:"id"`
	OrderID               uint                      `json:"order_id"`
	Status                string                    `json:"status"`
	ResponseBy            string                    `json:"response_by"`
	ResponseTime          *time.Time                `json:"response_time"`
	MarketingResponseBy   string                    `json:"marketing_response_by"`
	MarketingResponseTime *time.Time                `json:"marketing_response_time"`
	RevisiGeneral         string                    `json:"revisi_general"`
	CreatedAt             time.Time                 `json:"created_at"`
	UpdatedAt             time.Time                 `json:"updated_at"`
	Order                 *OrderBriefResponse       `json:"order,omitempty"`
	Files                 []GambarKerjaFileResponse `json:"files"`
}
