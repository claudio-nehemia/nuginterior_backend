package dto

import "time"

// Request DTOs
type UpdateApprovalMaterialItemRequest struct {
	ID           uint     `json:"id" binding:"required"`
	Area         string   `json:"area"`
	Foto         string   `json:"foto"`
	KodeMaterial []string `json:"kode_material"`
	BrandSpek    []string `json:"brand_spek"`
	Notes        string   `json:"notes"`
}

type UpdateApprovalMaterialRequest struct {
	Status string                              `json:"status"` // pending, completed
	Items  []UpdateApprovalMaterialItemRequest `json:"items"`
}

// Response DTOs
type ApprovalMaterialItemResponse struct {
	ID                 uint      `json:"id"`
	ApprovalMaterialID uint      `json:"approval_material_id"`
	Category           string    `json:"category"`
	SourceID           uint      `json:"source_id"`
	ItemName           string    `json:"item_name"`
	Area               string    `json:"area"`
	Foto               string    `json:"foto"`
	KodeMaterial       []string  `json:"kode_material"`
	BrandSpek          []string  `json:"brand_spek"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ApprovalMaterialResponse struct {
	ID                    uint                           `json:"id"`
	OrderID               uint                           `json:"order_id"`
	Status                string                         `json:"status"`
	ResponseBy            string                         `json:"response_by"`
	ResponseTime          *time.Time                     `json:"response_time"`
	MarketingResponseBy   string                         `json:"marketing_response_by"`
	MarketingResponseTime *time.Time                     `json:"marketing_response_time"`
	CreatedAt             time.Time                      `json:"created_at"`
	UpdatedAt             time.Time                      `json:"updated_at"`
	Order                 *OrderBriefResponse            `json:"order,omitempty"`
	Items                 []ApprovalMaterialItemResponse `json:"items"`
}
