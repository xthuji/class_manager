package services

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
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
	subjectsJSON, _ := json.Marshal(req.Subjects)

	studentID := req.StudentID
	if studentID == "" {
		var err error
		studentID, err = s.GenerateStudentID()
		if err != nil {
			return nil, err
		}
	}

	query := `INSERT INTO students (name, student_id, contact, subjects, total_hours, completed_hours, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.DB.Exec(query, req.Name, studentID, req.Contact, string(subjectsJSON), req.TotalHours, 0, timestamp, timestamp)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetStudentByID(id)
}

func (s *StudentService) UpdateStudent(req models.StudentUpdateRequest) (*models.Student, error) {
	timestamp := db.GetTimestamp()
	subjectsJSON, _ := json.Marshal(req.Subjects)

	query := `UPDATE students SET name=?, contact=?, subjects=?, total_hours=?, updated_at=? WHERE id=?`

	_, err := db.DB.Exec(query, req.Name, req.Contact, string(subjectsJSON), req.TotalHours, timestamp, req.ID)
	if err != nil {
		return nil, err
	}

	return s.GetStudentByID(req.ID)
}

func (s *StudentService) DeleteStudent(id int64) error {
	query := `DELETE FROM students WHERE id=?`
	_, err := db.DB.Exec(query, id)
	return err
}

func (s *StudentService) GetStudentByID(id int64) (*models.Student, error) {
	query := `SELECT id, name, student_id, contact, subjects, total_hours, completed_hours, created_at, updated_at FROM students WHERE id=?`

	row := db.DB.QueryRow(query, id)
	return s.scanStudent(row)
}

func (s *StudentService) ListStudents(req models.StudentListRequest) ([]models.Student, error) {
	offset := (req.Page - 1) * req.PageSize
	query := `SELECT id, name, student_id, contact, subjects, total_hours, completed_hours, created_at, updated_at FROM students LIMIT ? OFFSET ?`

	rows, err := db.DB.Query(query, req.PageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanStudents(rows)
}

func (s *StudentService) SearchStudents(req models.StudentSearchRequest) ([]models.Student, error) {
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

	if len(req.Subjects) > 0 {
		orConditions := make([]string, len(req.Subjects))
		for i, subject := range req.Subjects {
			orConditions[i] = "subjects LIKE ?"
			args = append(args, "%"+subject+"%")
		}
		conditions = append(conditions, "("+strings.Join(orConditions, " OR ")+")")
	}

	query := `SELECT id, name, student_id, contact, subjects, total_hours, completed_hours, created_at, updated_at FROM students`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanStudents(rows)
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
		subjectsStr := studentData["科目"]
		totalHoursStr := studentData["初始课时"]

		if name == "" || studentID == "" {
			failedCount++
			errors = append(errors, fmt.Sprintf("姓名或学号为空: %s", studentID))
			continue
		}

		var totalHours float64 = 0
		if totalHoursStr != "" {
			totalHours, err = strconv.ParseFloat(totalHoursStr, 64)
			if err != nil {
				failedCount++
				errors = append(errors, fmt.Sprintf("课时格式错误: %s", studentID))
				continue
			}
		}

		subjects := strings.Split(subjectsStr, ",")
		for i := range subjects {
			subjects[i] = strings.TrimSpace(subjects[i])
		}

		req := models.StudentCreateRequest{
			Name:       name,
			StudentID:  studentID,
			Contact:    contact,
			Subjects:   subjects,
			TotalHours: totalHours,
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
		query = fmt.Sprintf(`SELECT name, student_id, contact, subjects, total_hours, completed_hours FROM students WHERE id IN (%s)`, strings.Join(placeholders, ","))
	} else {
		query = `SELECT name, student_id, contact, subjects, total_hours, completed_hours FROM students`
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("姓名,学号,联系方式,科目,总课时,已完成课时\n")

	for rows.Next() {
		var name, studentID, contact, subjectsStr string
		var totalHours, completedHours float64

		if err := rows.Scan(&name, &studentID, &contact, &subjectsStr, &totalHours, &completedHours); err != nil {
			return "", err
		}

		var subjects []string
		json.Unmarshal([]byte(subjectsStr), &subjects)

		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%.2f,%.2f\n",
			name, studentID, contact, strings.Join(subjects, ","), totalHours, completedHours))
	}

	return sb.String(), nil
}

func (s *StudentService) scanStudent(row *sql.Row) (*models.Student, error) {
	var student models.Student
	var subjectsStr string
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&student.ID,
		&student.Name,
		&student.StudentID,
		&student.Contact,
		&subjectsStr,
		&student.TotalHours,
		&student.CompletedHours,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(subjectsStr), &student.Subjects)
	student.RemainingHours = student.TotalHours - student.CompletedHours
	student.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	student.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &student, nil
}

func (s *StudentService) scanStudents(rows *sql.Rows) ([]models.Student, error) {
	var students []models.Student

	for rows.Next() {
		var student models.Student
		var subjectsStr string
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&student.ID,
			&student.Name,
			&student.StudentID,
			&student.Contact,
			&subjectsStr,
			&student.TotalHours,
			&student.CompletedHours,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(subjectsStr), &student.Subjects)
		student.RemainingHours = student.TotalHours - student.CompletedHours
		student.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		student.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		students = append(students, student)
	}

	return students, nil
}
