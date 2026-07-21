package models

import "time"

type Schedule struct {
	ID            int64     `json:"id"`
	CourseID      int64     `json:"course_id"`
	DayOfWeek     int       `json:"day_of_week"`
	Period        string    `json:"period"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	HoursConsumed float64   `json:"hours_consumed"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ScheduleCreateRequest struct {
	CourseID      int64   `json:"course_id"`
	DayOfWeek     int     `json:"day_of_week"`
	Period        string  `json:"period"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	HoursConsumed float64 `json:"hours_consumed"`
}

type ScheduleUpdateRequest struct {
	ID            int64   `json:"id"`
	DayOfWeek     int     `json:"day_of_week"`
	Period        string  `json:"period"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	HoursConsumed float64 `json:"hours_consumed"`
}
