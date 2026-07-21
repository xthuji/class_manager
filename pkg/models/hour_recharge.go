package models

import "time"

type HourRecharge struct {
	ID            int64     `json:"id"`
	StudentID     int64     `json:"student_id"`
	StudentName   string    `json:"student_name"`
	StudentNo     string    `json:"student_no"`
	Hours         float64   `json:"hours"`
	RechargeDate  string    `json:"recharge_date"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type HourRechargeCreateRequest struct {
	StudentID   int64   `json:"student_id"`
	Hours       float64 `json:"hours"`
	RechargeDate string `json:"recharge_date"`
	Description string  `json:"description"`
}

type HourRechargeListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	StudentID int64 `json:"student_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}