package models

import "time"

type OperationLog struct {
	ID            int64     `json:"id"`
	OperationType string    `json:"operation_type"` // create, update, delete
	EntityType    string    `json:"entity_type"`    // student, course, hour_record, hour_recharge, schedule, notification
	EntityID      int64     `json:"entity_id"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type OperationLogListRequest struct {
	Page           int      `json:"page"`
	PageSize       int      `json:"page_size"`
	EntityType     string   `json:"entity_type"`
	OperationType  string   `json:"operation_type"`
	EntityTypes    []string `json:"entity_types"`
	OperationTypes []string `json:"operation_types"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
}
