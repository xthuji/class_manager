package models

import "time"

type Threshold struct {
	ID           int64     `json:"id"`
	Subject      string    `json:"subject"`
	Value        float64   `json:"value"`
	TemplateName string    `json:"template_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ThresholdRequest struct {
	Subject string  `json:"subject"`
	Value   float64 `json:"value"`
}

type ThresholdTemplateRequest struct {
	TemplateName string             `json:"template_name"`
	Thresholds   []ThresholdRequest `json:"thresholds"`
}

type ThresholdPreviewResponse struct {
	Subject         string `json:"subject"`
	ThresholdValue  float64 `json:"threshold_value"`
	TriggerCount    int     `json:"trigger_count"`
	TotalStudentCount int   `json:"total_student_count"`
}
