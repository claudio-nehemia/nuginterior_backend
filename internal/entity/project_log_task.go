package entity

import (
	"time"
)

// ProjectLogTask represents the log of task activities for project stages.
type ProjectLogTask struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID      uint       `gorm:"not null;index" json:"order_id"`
	Order        *Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Stage        string     `gorm:"size:50;not null" json:"stage"`
	CreatedAt    time.Time  `json:"created_at"`           // Transition time
	TouchedAt    *time.Time `json:"touched_at,omitempty"` // When data was first saved/uploaded
	TouchedBy    string     `gorm:"size:255" json:"touched_by,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"` // When stage completed (next stage start)
	CompletedBy  string     `gorm:"size:255" json:"completed_by,omitempty"`
	DeadlineDays int        `gorm:"not null;default:0" json:"deadline_days"`
}

func (ProjectLogTask) TableName() string {
	return "project_log_tasks"
}
