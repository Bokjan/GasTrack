package dto

import "time"

// --- 通知相关 DTO ---

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID         string    `json:"id"`
	VehicleID  *string   `json:"vehicle_id,omitempty"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	ReminderID *string   `json:"reminder_id,omitempty"`
	RecordID   *string   `json:"record_id,omitempty"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

// UnreadCountResponse 未读通知数响应
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
