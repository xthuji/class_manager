package models

import "time"

type Course struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Subject      string    `json:"subject"`
	Description  string    `json:"description"`
	StudentCount int       `json:"student_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CourseCreateRequest struct {
	Name        string `json:"name"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

type CourseUpdateRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

type CourseListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Subject  string `json:"subject"`
}

type EnrollRequest struct {
	StudentID int64 `json:"student_id"`
	CourseID  int64 `json:"course_id"`
}

type CourseHoursRequest struct {
	CourseID int64   `json:"course_id"`
	Hours    float64 `json:"hours"`
}