package models

import "time"

type Notification struct {
	ID           int64     `json:"id"`
	StudentID    int64     `json:"student_id"`
	StudentName  string    `json:"student_name"`
	Subject      string    `json:"subject"`
	CurrentHours float64   `json:"current_hours"`
	Threshold    float64   `json:"threshold"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	ProcessedAt  time.Time `json:"processed_at"`
}

type NotificationRequest struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Body     string `json:"body"`
}
