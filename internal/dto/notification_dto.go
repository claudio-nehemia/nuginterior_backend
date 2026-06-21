package dto

import "time"

type NotificationResponse struct {
	ID          uint      `json:"id"`
	OrderID     *uint     `json:"order_id,omitempty"`
	NomorOrder  string    `json:"nomor_order,omitempty"`
	NamaProject string    `json:"nama_project,omitempty"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Link        string    `json:"link,omitempty"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type NotificationUnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}
