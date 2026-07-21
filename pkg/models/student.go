package models

import "time"

type Student struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	StudentID      string    `json:"student_id"`
	Contact        string    `json:"contact"`
	Subjects       []string  `json:"subjects"`
	TotalHours     float64   `json:"total_hours"`
	CompletedHours float64   `json:"completed_hours"`
	RemainingHours float64   `json:"remaining_hours"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StudentCreateRequest struct {
	Name           string   `json:"name"`
	StudentID      string   `json:"student_id"`
	Contact        string   `json:"contact"`
	Subjects       []string `json:"subjects"`
	TotalHours     float64  `json:"total_hours"`
}

type StudentUpdateRequest struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Contact        string    `json:"contact"`
	Subjects       []string  `json:"subjects"`
	TotalHours     float64   `json:"total_hours"`
}

type StudentListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type StudentSearchRequest struct {
	Name       string   `json:"name"`
	StudentID  string   `json:"student_id"`
	Subjects   []string `json:"subjects"`
}

type BatchImportResult struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	Errors       []string `json:"errors"`
}
