package dto

import "time"

type CreateDesainFinalRequest struct {
	OrderID uint `json:"order_id" binding:"required"`
}

type AcceptDesainFinalRequest struct {
	DesainFinalFileID uint `json:"desain_final_file_id" binding:"required"`
}

type ReviseDesainFinalRequest struct {
	DesainFinalFileID uint   `json:"desain_final_file_id" binding:"required"`
	Notes             string `json:"notes" binding:"required"`
}

type DesainFinalFileResponse struct {
	ID            uint      `json:"id"`
	DesainFinalID uint      `json:"desain_final_id"`
	FilePath      string    `json:"file_path"`
	OriginalName  string    `json:"original_name"`
	Status        string    `json:"status"`
	Revisi        string    `json:"revisi"`
	CreatedAt     time.Time `json:"created_at"`
}

type DesainFinalResponse struct {
	ID                    uint                      `json:"id"`
	OrderID               uint                      `json:"order_id"`
	Status                string                    `json:"status"`
	ResponseTime          *time.Time                `json:"response_time"`
	ResponseBy            string                    `json:"response_by"`
	MarketingResponseTime *time.Time                `json:"marketing_response_time"`
	MarketingResponseBy   string                    `json:"marketing_response_by"`
	CreatedAt             time.Time                 `json:"created_at"`
	UpdatedAt             time.Time                 `json:"updated_at"`
	Order                 *OrderBriefResponse       `json:"order,omitempty"`
	Files                 []DesainFinalFileResponse `json:"files"`
}
