package models

import "time"

type Notification struct {
	ID           int64     `json:"id"`
	StudentID    int64     `json:"student_id"`
	StudentName  string    `json:"student_name"`
	CourseName   string    `json:"course_name"`
	CurrentHours float64   `json:"current_hours"`
	Threshold    float64   `json:"threshold"`
	Status       string    `json:"status"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"created_at"`
}

type NotificationRequest struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Body     string `json:"body"`
}
