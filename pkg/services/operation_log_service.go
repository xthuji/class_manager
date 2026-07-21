package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/utils"
)

type OperationLogService struct{}

// FieldChange represents a single field change between an old and new value.
type FieldChange struct {
	Field string      // field display name in Chinese
	Old   interface{} // previous value
	New   interface{} // new value
}

func NewOperationLogService() *OperationLogService {
	return &OperationLogService{}
}

// Log creates a new operation log entry.
func (s *OperationLogService) Log(operationType, entityType string, entityID int64, description string) error {
	timestamp := db.GetTimestamp()

	query := `INSERT INTO operation_logs (operation_type, entity_type, entity_id, description, created_at)
			  VALUES (?, ?, ?, ?, ?)`

	_, err := db.DB.Exec(query, operationType, entityType, entityID, description, timestamp)
	if err != nil {
		utils.LogErrorf("Failed to log operation: %v", err)
		return err
	}

	return nil
}

// getEntityLabel maps an entity type to its Chinese display label.
func getEntityLabel(entityType string) string {
	switch entityType {
	case "student":
		return "学生"
	case "course":
		return "课程"
	case "hour_record":
		return "课时记录"
	case "hour_recharge":
		return "充值"
	case "schedule":
		return "课程" // 向后兼容：旧记录映射到课程
	case "notification":
		return "通知"
	default:
		return entityType
	}
}

// formatValue formats a field value for display in the log description.
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case float64:
		return fmt.Sprintf("%.1f", val)
	case float32:
		return fmt.Sprintf("%.1f", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// resolveStudentName looks up a student's name by ID for log display.
func resolveStudentName(id int64) string {
	var name string
	err := db.DB.QueryRow("SELECT name FROM students WHERE id=?", id).Scan(&name)
	if err != nil {
		return fmt.Sprintf("ID:%d", id)
	}
	return name
}

// resolveCourseName looks up a course's name by ID for log display.
func resolveCourseName(id int64) string {
	var name string
	err := db.DB.QueryRow("SELECT name FROM courses WHERE id=?", id).Scan(&name)
	if err != nil {
		return fmt.Sprintf("ID:%d", id)
	}
	return name
}

// resolveStudentNames looks up multiple student names by IDs for log display.
func resolveStudentNames(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := "SELECT id, name FROM students WHERE id IN (" + strings.Join(placeholders, ",") + ")"
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return fmt.Sprintf("%d students", len(ids))
	}
	defer rows.Close()
	names := make([]string, 0, len(ids))
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return fmt.Sprintf("%d students", len(ids))
	}
	return strings.Join(names, ", ")
}

// LogChange records an operation log with structured field changes.
// For create operations, entityName should identify the entity; changes is ignored.
// For update operations, only changed fields should be passed; if changes is empty, no log is created.
// For delete operations, entityName should be the entity's display name.
func (s *OperationLogService) LogChange(operationType, entityType string, entityID int64, entityName string, changes []FieldChange) error {
	label := getEntityLabel(entityType)
	var description string

	switch operationType {
	case "create":
		description = fmt.Sprintf("创建%s: %s", label, entityName)
	case "update":
		if len(changes) == 0 {
			return nil
		}
		parts := make([]string, 0, len(changes))
		for _, c := range changes {
			parts = append(parts, fmt.Sprintf("%s: %s→%s", c.Field, formatValue(c.Old), formatValue(c.New)))
		}
		description = fmt.Sprintf("修改%s: %s | %s", label, entityName, strings.Join(parts, ", "))
	case "delete":
		if entityName != "" {
			description = fmt.Sprintf("删除%s: %s (ID: %d)", label, entityName, entityID)
		} else {
			description = fmt.Sprintf("删除%s (ID: %d)", label, entityID)
		}
	default:
		description = fmt.Sprintf("%s%s: %s", operationType, label, entityName)
	}

	return s.Log(operationType, entityType, entityID, description)
}

// ListLogs lists operation logs with filtering and pagination.
func (s *OperationLogService) ListLogs(req models.OperationLogListRequest) (models.PaginatedResponse, error) {
	query := `SELECT id, operation_type, entity_type, entity_id, description, created_at FROM operation_logs WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM operation_logs WHERE 1=1`

	var args []interface{}

	if len(req.EntityTypes) > 0 {
		placeholders := make([]string, len(req.EntityTypes))
		for i, v := range req.EntityTypes {
			placeholders[i] = "?"
			args = append(args, v)
		}
		query += " AND entity_type IN (" + strings.Join(placeholders, ",") + ")"
		countQuery += " AND entity_type IN (" + strings.Join(placeholders, ",") + ")"
	} else if req.EntityType != "" {
		query += " AND entity_type = ?"
		countQuery += " AND entity_type = ?"
		args = append(args, req.EntityType)
	}

	if len(req.OperationTypes) > 0 {
		placeholders := make([]string, len(req.OperationTypes))
		for i, v := range req.OperationTypes {
			placeholders[i] = "?"
			args = append(args, v)
		}
		query += " AND operation_type IN (" + strings.Join(placeholders, ",") + ")"
		countQuery += " AND operation_type IN (" + strings.Join(placeholders, ",") + ")"
	} else if req.OperationType != "" {
		query += " AND operation_type = ?"
		countQuery += " AND operation_type = ?"
		args = append(args, req.OperationType)
	}

	if req.StartDate != "" {
		query += " AND created_at >= ?"
		countQuery += " AND created_at >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND created_at <= ?"
		countQuery += " AND created_at <= ?"
		args = append(args, req.EndDate+"T23:59:59")
	}

	// 获取总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	var total int
	err := db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	query += " ORDER BY id DESC"

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	query += " LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	defer rows.Close()

	logs := []models.OperationLog{}
	for rows.Next() {
		var logEntry models.OperationLog
		var createdAtStr string
		var description sql.NullString

		err := rows.Scan(
			&logEntry.ID,
			&logEntry.OperationType,
			&logEntry.EntityType,
			&logEntry.EntityID,
			&description,
			&createdAtStr,
		)
		if err != nil {
			return models.PaginatedResponse{}, err
		}

		logEntry.Description = description.String
		logEntry.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		logs = append(logs, logEntry)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return models.PaginatedResponse{
		Data:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// DeleteLog deletes a single operation log entry by ID.
func (s *OperationLogService) DeleteLog(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM operation_logs WHERE id = ?`, id)
	return err
}

// BatchDeleteLogs 批量删除操作日志
func (s *OperationLogService) BatchDeleteLogs(ids []int64) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.DeleteLog(id); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
