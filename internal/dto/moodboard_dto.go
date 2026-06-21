package dto

import "time"

// Moodboard Request DTOs

type CreateMoodboardRequest struct {
	OrderID uint `json:"order_id" binding:"required"`
}

type UpdateMoodboardRequest struct {
	Notes       string `json:"notes"`
	RevisiFinal string `json:"revisi_final"`
	Status      string `json:"status"`
}

type AcceptDesainRequest struct {
	MoodboardFileID uint `json:"moodboard_file_id" binding:"required"`
}

type ReviseRequest struct {
	MoodboardFileID uint   `json:"moodboard_file_id" binding:"required"`
	Notes           string `json:"notes" binding:"required"`
}

type ReviseFinalRequest struct {
	MoodboardFileID uint   `json:"moodboard_file_id" binding:"required"`
	Notes           string `json:"notes" binding:"required"`
}

type UpdateTotalFeeRequest struct {
	TotalFee float64 `json:"total_fee" binding:"required,gt=0"`
}

// Moodboard Response DTOs

type MoodboardFileResponse struct {
	ID           uint      `json:"id"`
	FilePath     string    `json:"file_path"`
	FileType     string    `json:"file_type"`
	OriginalName string    `json:"original_name"`
	Status       string    `json:"status"`
	Revisi       string    `json:"revisi"`
	CreatedAt    time.Time `json:"created_at"`
}

type EstimasiFileResponse struct {
	ID              uint      `json:"id"`
	MoodboardFileID uint      `json:"moodboard_file_id"`
	FilePath        string    `json:"file_path"`
	OriginalName    string    `json:"original_name"`
	CreatedAt       time.Time `json:"created_at"`
}

type EstimasiResponse struct {
	ID             uint                   `json:"id"`
	EstimatedCost  string                 `json:"estimated_cost"`
	ResponseBy     string                 `json:"response_by"`
	ResponseTime   *time.Time             `json:"response_time"`
	PmResponseBy   string                 `json:"pm_response_by"`
	PmResponseTime *time.Time             `json:"pm_response_time"`
	Files          []EstimasiFileResponse `json:"files"`
}

type CommitmentFeeResponse struct {
	ID             uint       `json:"id"`
	TotalFee       *float64   `json:"total_fee"`
	PaymentProof   string     `json:"payment_proof"`
	PaymentStatus  string     `json:"payment_status"`
	ResponseBy     string     `json:"response_by"`
	ResponseTime   *time.Time `json:"response_time"`
	PmResponseBy   string     `json:"pm_response_by"`
	PmResponseTime *time.Time `json:"pm_response_time"`
}

type MoodboardResponse struct {
	ID                  uint                    `json:"id"`
	OrderID             uint                    `json:"order_id"`
	MoodboardKasar      string                  `json:"moodboard_kasar"`
	MoodboardFinal      string                  `json:"moodboard_final"`
	Status                string                  `json:"status"`
	Notes                 string                  `json:"notes"`
	RevisiFinal           string                  `json:"revisi_final"`
	ResponseTime          *time.Time              `json:"response_time"`
	ResponseBy            string                  `json:"response_by"`
	MarketingResponse     string                  `json:"marketing_response"`
	MarketingResponseBy   string                  `json:"marketing_response_by"`
	MarketingResponseTime *time.Time              `json:"marketing_response_time"`
	CreatedAt             time.Time               `json:"created_at"`
	UpdatedAt             time.Time               `json:"updated_at"`
	Order                 *OrderBriefResponse     `json:"order,omitempty"`
	Files                 []MoodboardFileResponse `json:"files"`
	Estimasi              *EstimasiResponse       `json:"estimasi,omitempty"`
	CommitmentFee         *CommitmentFeeResponse  `json:"commitment_fee,omitempty"`
}
