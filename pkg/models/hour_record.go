package models

import "time"

type HourRecord struct {
	ID         int64     `json:"id"`
	StudentID  int64     `json:"student_id"`
	CourseID   int64     `json:"course_id"`
	Hours      float64   `json:"hours"`
	RecordDate string    `json:"record_date"`
	Description string   `json:"description"`
	CreatedAt  time.Time `json:"created_at"`
}

type HourRecordCreateRequest struct {
	StudentID   int64   `json:"student_id"`
	CourseID    int64   `json:"course_id"`
	Hours       float64 `json:"hours"`
	RecordDate  string  `json:"record_date"`
	Description string  `json:"description"`
}

type BatchHourRecordRequest struct {
	StudentIDs  []int64  `json:"student_ids"`
	CourseID    int64    `json:"course_id"`
	Hours       float64  `json:"hours"`
	RecordDate  string   `json:"record_date"`
	Description string   `json:"description"`
}

type HoursSummaryResponse struct {
	TotalHours     float64 `json:"total_hours"`
	CompletedHours float64 `json:"completed_hours"`
	RemainingHours float64 `json:"remaining_hours"`
}

type HourRecordListRequest struct {
	StudentIDs []int64 `json:"student_ids"`
	CourseIDs  []int64 `json:"course_ids"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
}

type HourRecordWithDetails struct {
	ID            int64     `json:"id"`
	StudentID     int64     `json:"student_id"`
	CourseID      int64     `json:"course_id"`
	Hours         float64   `json:"hours"`
	RecordDate    string    `json:"record_date"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	StudentName   string    `json:"student_name"`
	StudentNo     string    `json:"student_no"`
	CourseName    string    `json:"course_name"`
	CourseSubject string    `json:"course_subject"`
}

type ExportRequest struct {
	StudentID  int64  `json:"student_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}