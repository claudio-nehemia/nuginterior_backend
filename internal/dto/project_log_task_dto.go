package dto

import (
	"fmt"
	"strings"
	"time"
)

type ProjectLogTaskResponse struct {
	ID                 uint       `json:"id"`
	OrderID            uint       `json:"order_id"`
	NomorOrder         string     `json:"nomor_order"`
	NamaProject        string     `json:"nama_project"`
	NamaCustomer       string     `json:"nama_customer"`
	Stage              string     `json:"stage"`
	StageLabel         string     `json:"stage_label"`
	CreatedAt          time.Time  `json:"created_at"` // Transition time
	TouchedAt          *time.Time `json:"touched_at,omitempty"`
	TouchedBy          string     `json:"touched_by,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CompletedBy        string     `json:"completed_by,omitempty"`
	DeadlineDays       int        `json:"deadline_days"`
	DeadlineTime       time.Time  `json:"deadline_time"`
	IsLate             bool       `json:"is_late"`
	LateDays           int        `json:"late_days"`
	DurationToTouch    string     `json:"duration_to_touch"`
	DurationToComplete string     `json:"duration_to_complete"`
}

// FormatDurationIndonesian formats a duration into human-readable Indonesian terms.
func FormatDurationIndonesian(d time.Duration) string {
	if d < 0 {
		return "-"
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d hari", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d jam", hours))
	}
	if minutes > 0 || (days == 0 && hours == 0) {
		parts = append(parts, fmt.Sprintf("%d menit", minutes))
	}

	return strings.Join(parts, " ")
}
