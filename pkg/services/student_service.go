package services

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/utils"
)

type StudentService struct{}

func NewStudentService() *StudentService {
	return &StudentService{}
}

func (s *StudentService) GenerateStudentID() (string, error) {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("STU%04d", count+1), nil
}

func (s *StudentService) CreateStudent(req models.StudentCreateRequest) (*models.Student, error) {
	timestamp := db.GetTimestamp()

	studentID := req.StudentID
	if studentID == "" {
		var err error
		studentID, err = s.GenerateStudentID()
		if err != nil {
			return nil, err
		}
	}

	query := `INSERT INTO students (name, student_id, contact, total_hours, completed_hours, is_dropped, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.DB.Exec(query, req.Name, studentID, req.Contact, 0, 0, 0, timestamp, timestamp)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	utils.LogInfof("学生创建: ID=%d, 姓名=%s, 学号=%s", id, req.Name, studentID)
	logService := NewOperationLogService()
	logService.LogChange("create", "student", id, fmt.Sprintf("%s (%s)", req.Name, studentID), nil)
	return s.GetStudentByID(id)
}

func (s *StudentService) UpdateStudent(req models.StudentUpdateRequest) (*models.Student, error) {
	old, _ := s.GetStudentByID(req.ID)

	timestamp := db.GetTimestamp()

	query := `UPDATE students SET name=?, contact=?, total_hours=?, is_dropped=?, updated_at=? WHERE id=?`

	isDropped := 0
	if req.IsDropped {
		isDropped = 1
	}
	_, err := db.DB.Exec(query, req.Name, req.Contact, req.TotalHours, isDropped, timestamp, req.ID)
	if err != nil {
		return nil, err
	}

	utils.LogInfof("学生更新: ID=%d, 姓名=%s", req.ID, req.Name)
	logService := NewOperationLogService()
	var changes []FieldChange
	if old != nil {
		if old.Name != req.Name {
			changes = append(changes, FieldChange{Field: "姓名", Old: old.Name, New: req.Name})
		}
		if old.Contact != req.Contact {
			changes = append(changes, FieldChange{Field: "联系方式", Old: old.Contact, New: req.Contact})
		}
		if old.TotalHours != req.TotalHours {
			changes = append(changes, FieldChange{Field: "总课时", Old: old.TotalHours, New: req.TotalHours})
		}
		if old.IsDropped != req.IsDropped {
			changes = append(changes, FieldChange{Field: "退学状态", Old: old.IsDropped, New: req.IsDropped})
		}
	}
	logService.LogChange("update", "student", req.ID, req.Name, changes)
	return s.GetStudentByID(req.ID)
}

func (s *StudentService) DeleteStudent(id int64) error {
	old, _ := s.GetStudentByID(id)

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	// 删除关联数据（无外键级联，需手动清理）
	if _, err := tx.Exec(`DELETE FROM student_course WHERE student_id=?`, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`DELETE FROM hour_records WHERE student_id=?`, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`DELETE FROM hour_recharges WHERE student_id=?`, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`DELETE FROM notifications WHERE student_id=?`, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`DELETE FROM students WHERE id=?`, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	utils.LogInfof("学生删除: ID=%d", id)
	logService := NewOperationLogService()
	var entityName string
	if old != nil {
		entityName = old.Name
	}
	logService.LogChange("delete", "student", id, entityName, nil)
	return nil
}

// BatchDeleteStudents 批量删除学生
func (s *StudentService) BatchDeleteStudents(ids []int64) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.DeleteStudent(id); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *StudentService) GetStudentByID(id int64) (*models.Student, error) {
	query := `SELECT id, name, student_id, contact, total_hours, completed_hours, is_dropped, created_at, updated_at FROM students WHERE id=?`

	row := db.DB.QueryRow(query, id)
	return s.scanStudent(row)
}

func (s *StudentService) ListStudents(req models.StudentListRequest) (models.PaginatedResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	
	// 获取总数
	var total int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&total)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	
	query := `SELECT id, name, student_id, contact, total_hours, completed_hours, is_dropped, created_at, updated_at FROM students ORDER BY id ASC LIMIT ? OFFSET ?`

	rows, err := db.DB.Query(query, pageSize, offset)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	defer rows.Close()

	students, err := s.scanStudents(rows)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	
	totalPages := (total + pageSize - 1) / pageSize

	return models.PaginatedResponse{
		Data:       students,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *StudentService) SearchStudents(req models.StudentSearchRequest) (models.PaginatedResponse, error) {
	var conditions []string
	var args []interface{}

	if req.Name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+req.Name+"%")
	}

	if req.StudentID != "" {
		conditions = append(conditions, "student_id LIKE ?")
		args = append(args, "%"+req.StudentID+"%")
	}

	if req.IsDropped != nil {
		if *req.IsDropped {
			conditions = append(conditions, "is_dropped = 1")
		} else {
			conditions = append(conditions, "is_dropped = 0")
		}
	}

	query := `SELECT id, name, student_id, contact, total_hours, completed_hours, is_dropped, created_at, updated_at FROM students`
	countQuery := `SELECT COUNT(*) FROM students`
	
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	// 获取总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	var total int
	err := db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	query += " ORDER BY id ASC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	defer rows.Close()

	students, err := s.scanStudents(rows)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	
	totalPages := (total + pageSize - 1) / pageSize

	return models.PaginatedResponse{
		Data:       students,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *StudentService) BatchImportStudents(csvContent string) (models.BatchImportResult, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))
	header, err := reader.Read()
	if err != nil {
		return models.BatchImportResult{}, err
	}

	successCount := 0
	failedCount := 0
	var errors []string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return models.BatchImportResult{}, err
		}

		studentData := make(map[string]string)
		for i, field := range header {
			if i < len(record) {
				studentData[strings.TrimSpace(field)] = strings.TrimSpace(record[i])
			}
		}

		name := studentData["姓名"]
		studentID := studentData["学号"]
		contact := studentData["联系方式"]

		if name == "" || studentID == "" {
			failedCount++
			errors = append(errors, fmt.Sprintf("姓名或学号为空: %s", studentID))
			continue
		}

		req := models.StudentCreateRequest{
			Name:      name,
			StudentID: studentID,
			Contact:   contact,
		}

		if _, err := s.CreateStudent(req); err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("创建学生失败 %s: %v", studentID, err))
		} else {
			successCount++
		}
	}

	return models.BatchImportResult{
		SuccessCount: successCount,
		FailedCount:  failedCount,
		Errors:       errors,
	}, nil
}

func (s *StudentService) ExportStudents(ids []int64) (string, error) {
	var query string
	var args []interface{}

	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		for i := range placeholders {
			placeholders[i] = "?"
			args = append(args, ids[i])
		}
		query = fmt.Sprintf(`SELECT name, student_id, contact, total_hours, completed_hours FROM students WHERE id IN (%s)`, strings.Join(placeholders, ","))
	} else {
		query = `SELECT name, student_id, contact, total_hours, completed_hours FROM students`
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("姓名,学号,联系方式,总课时,已完成课时\n")

	for rows.Next() {
		var name, studentID, contact string
		var totalHours, completedHours float64

		if err := rows.Scan(&name, &studentID, &contact, &totalHours, &completedHours); err != nil {
			return "", err
		}

		sb.WriteString(fmt.Sprintf("%s,%s,%s,%.2f,%.2f\n",
			name, studentID, contact, totalHours, completedHours))
	}

	return sb.String(), nil
}

func (s *StudentService) scanStudent(row *sql.Row) (*models.Student, error) {
	var student models.Student
	var createdAtStr, updatedAtStr string
	var isDropped int

	err := row.Scan(
		&student.ID,
		&student.Name,
		&student.StudentID,
		&student.Contact,
		&student.TotalHours,
		&student.CompletedHours,
		&isDropped,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	student.IsDropped = isDropped == 1
	student.RemainingHours = student.TotalHours - student.CompletedHours
	student.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	student.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &student, nil
}

func (s *StudentService) scanStudents(rows *sql.Rows) ([]models.Student, error) {
	students := []models.Student{}

	for rows.Next() {
		var student models.Student
		var createdAtStr, updatedAtStr string
		var isDropped int

		err := rows.Scan(
			&student.ID,
			&student.Name,
			&student.StudentID,
			&student.Contact,
			&student.TotalHours,
			&student.CompletedHours,
			&isDropped,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		student.IsDropped = isDropped == 1
		student.RemainingHours = student.TotalHours - student.CompletedHours
		student.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		student.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		students = append(students, student)
	}

	return students, nil
}
