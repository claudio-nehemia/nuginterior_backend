package dto

import "time"

type WorkplanStageMasterResponse struct {
	ID         uint      `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Percentage float64   `json:"percentage"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type WorkplanStageResponse struct {
	ID              uint                         `json:"id"`
	WorkplanID      uint                         `json:"workplan_id"`
	InputItemRoomID uint                         `json:"input_item_room_id"`
	StageMasterID   uint                         `json:"stage_master_id"`
	Percentage      float64                      `json:"percentage"`
	StartDate       *time.Time                   `json:"start_date"`
	EndDate         *time.Time                   `json:"end_date"`
	Notes           string                       `json:"notes"`
	Status          string                       `json:"status"`
	CompletedAt     *time.Time                   `json:"completed_at"`
	CompletedBy     string                       `json:"completed_by"`
	Photos          string                       `json:"photos"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
	RoomName        string                       `json:"room_name"`
	ProductName     string                       `json:"product_name"`
	ProductDims     string                       `json:"product_dims"`
	Qty             int                          `json:"qty"`
	StageMaster     *WorkplanStageMasterResponse `json:"stage_master,omitempty"`
}

type CompleteStageRequest struct {
	Photos []string `json:"photos" binding:"required,min=1"`
	Notes  string   `json:"notes"`
}

type RequestExtensionRequest struct {
	Notes string `json:"notes" binding:"required"`
	Days  int    `json:"days" binding:"required,min=1"`
}

type HandleExtensionRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
}

type WorkplanResponse struct {
	ID              uint                    `json:"id"`
	OrderID         uint                    `json:"order_id"`
	Status          string                  `json:"status"` // draft, submitted
	StartDate       *time.Time              `json:"start_date"`
	EndDate         *time.Time              `json:"end_date"`
	DurationDays    int                     `json:"duration_days"`
	ResponseBy      string                  `json:"response_by"`
	ResponseTime    *time.Time              `json:"response_time"`
	ExtensionStatus string                  `json:"extension_status"`
	ExtensionNotes  string                  `json:"extension_notes"`
	ExtensionDays   int                     `json:"extension_days"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	Order           *OrderBriefResponse     `json:"order,omitempty"`
	Stages          []WorkplanStageResponse `json:"stages,omitempty"`
	BastPhoto       string                  `json:"bast_photo"`
	BastGeneratedAt *time.Time              `json:"bast_generated_at"`
	BastGeneratedBy string                  `json:"bast_generated_by"`
	IsFullyPaid     bool                    `json:"is_fully_paid"`
}

type UpdateWorkplanStageRequest struct {
	InputItemRoomID uint    `json:"input_item_room_id" binding:"required"`
	StageMasterID   uint    `json:"stage_master_id" binding:"required"`
	Percentage      float64 `json:"percentage"`
	StartDate       *string `json:"start_date"` // YYYY-MM-DD
	EndDate         *string `json:"end_date"`   // YYYY-MM-DD
	Notes           string  `json:"notes"`
}

type UpdateWorkplanStageMasterRequest struct {
	ID         uint    `json:"id"`
	Code       string  `json:"code" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	Percentage float64 `json:"percentage" binding:"required,min=0,max=100"`
}

type UpdateWorkplanRequest struct {
	Status    string                       `json:"status" binding:"required,oneof=draft submitted"`
	StartDate *string                      `json:"start_date"` // YYYY-MM-DD
	EndDate   *string                      `json:"end_date"`   // YYYY-MM-DD
	Stages    []UpdateWorkplanStageRequest `json:"stages"`
}

// --- Defect DTOs ---

type ReportDefectRequest struct {
	Description string   `json:"description" binding:"required"`
	Photos      []string `json:"photos"`
}

type SubmitDefectFixRequest struct {
	FixDescription string   `json:"fix_description" binding:"required"`
	FixPhotos      []string `json:"fix_photos"`
}

type ReviewDefectRequest struct {
	Action      string `json:"action" binding:"required,oneof=accept reject"`
	ReviewNotes string `json:"review_notes"`
}

type WorkplanDefectResponse struct {
	ID              uint       `json:"id"`
	WorkplanStageID uint       `json:"workplan_stage_id"`
	Description     string     `json:"description"`
	Photos          string     `json:"photos"`
	Status          string     `json:"status"`
	FixDescription  string     `json:"fix_description"`
	FixPhotos       string     `json:"fix_photos"`
	ReportedBy      string     `json:"reported_by"`
	FixedBy         string     `json:"fixed_by"`
	ReviewedBy      string     `json:"reviewed_by"`
	ReviewNotes     string     `json:"review_notes"`
	ReportedAt      *time.Time `json:"reported_at"`
	FixedAt         *time.Time `json:"fixed_at"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	// Extra info from joined stage
	RoomName    string `json:"room_name,omitempty"`
	ProductName string `json:"product_name,omitempty"`
	StageName   string `json:"stage_name,omitempty"`
}

type SubmitBastRequest struct {
	BastPhoto string `json:"bast_photo" binding:"required"`
}


