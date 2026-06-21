package dto

import "time"

// Setting Request DTOs

type UpdateSettingRequest struct {
	Value string `json:"value" binding:"required"`
}

// Setting Response DTOs

type SettingResponse struct {
	ID          uint      `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
