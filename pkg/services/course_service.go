package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/utils"
)

type CourseService struct{}

func NewCourseService() *CourseService {
	return &CourseService{}
}

func (s *CourseService) CreateCourse(req models.CourseCreateRequest) (*models.Course, error) {
	// 检查课程名唯一性
	var existing int64
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM courses WHERE name=?`, req.Name).Scan(&existing)
	if err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, fmt.Errorf("课程名称 '%s' 已存在", req.Name)
	}

	timestamp := db.GetTimestamp()

	query := `INSERT INTO courses (name, description, threshold, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?)`

	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 2
	}

	result, err := db.DB.Exec(query, req.Name, req.Description, threshold, timestamp, timestamp)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	utils.LogInfof("课程创建: ID=%d, 名称=%s", id, req.Name)
	logService := NewOperationLogService()
	logService.LogChange("create", "course", id, req.Name, nil)
	return s.GetCourseByID(id)
}

func (s *CourseService) UpdateCourse(req models.CourseUpdateRequest) (*models.Course, error) {
	// 检查课程名唯一性（排除自身）
	var existing int64
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM courses WHERE name=? AND id!=?`, req.Name, req.ID).Scan(&existing)
	if err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, fmt.Errorf("课程名称 '%s' 已存在", req.Name)
	}

	old, _ := s.GetCourseByID(req.ID)

	timestamp := db.GetTimestamp()

	query := `UPDATE courses SET name=?, description=?, threshold=?, updated_at=? WHERE id=?`

	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 2
	}

	_, err = db.DB.Exec(query, req.Name, req.Description, threshold, timestamp, req.ID)
	if err != nil {
		return nil, err
	}

	utils.LogInfof("课程更新: ID=%d, 名称=%s", req.ID, req.Name)
	logService := NewOperationLogService()
	var changes []FieldChange
	if old != nil {
		if old.Name != req.Name {
			changes = append(changes, FieldChange{Field: "课程名称", Old: old.Name, New: req.Name})
		}
		if old.Description != req.Description {
			changes = append(changes, FieldChange{Field: "描述", Old: old.Description, New: req.Description})
		}
		if old.Threshold != threshold {
			changes = append(changes, FieldChange{Field: "阈值", Old: old.Threshold, New: threshold})
		}
	}
	logService.LogChange("update", "course", req.ID, req.Name, changes)
	return s.GetCourseByID(req.ID)
}

func (s *CourseService) DeleteCourse(id int64) error {
	old, _ := s.GetCourseByID(id)

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	// 删除关联数据（无外键级联，需手动清理）
	if _, err := tx.Exec(`DELETE FROM student_course WHERE course_id=?`, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`DELETE FROM schedules WHERE course_id=?`, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`DELETE FROM courses WHERE id=?`, id); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	utils.LogInfof("课程删除: ID=%d", id)
	logService := NewOperationLogService()
	var entityName string
	if old != nil {
		entityName = old.Name
	}
	logService.LogChange("delete", "course", id, entityName, nil)
	return nil
}

// BatchDeleteCourses 批量删除课程
func (s *CourseService) BatchDeleteCourses(ids []int64) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.DeleteCourse(id); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *CourseService) GetCourseByID(id int64) (*models.Course, error) {
	query := `SELECT c.id, c.name, c.description, c.threshold, c.created_at, c.updated_at,
		(SELECT COUNT(*) FROM student_course WHERE course_id = c.id) AS student_count
		FROM courses c WHERE c.id=?`

	row := db.DB.QueryRow(query, id)

	var course models.Course
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&course.ID,
		&course.Name,
		&course.Description,
		&course.Threshold,
		&createdAtStr,
		&updatedAtStr,
		&course.StudentCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	course.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	course.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &course, nil
}

func (s *CourseService) ListCourses(req models.CourseListRequest) (models.PaginatedResponse, error) {
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
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM courses`).Scan(&total)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	
	query := `SELECT c.id, c.name, c.description, c.threshold, c.created_at, c.updated_at,
		(SELECT COUNT(*) FROM student_course WHERE course_id = c.id) AS student_count,
		(SELECT COALESCE(SUM(s.total_hours - s.completed_hours), 0) FROM students s JOIN student_course sc ON s.id = sc.student_id WHERE sc.course_id = c.id) as remaining_hours
		FROM courses c`
	var args []interface{}

	query += " ORDER BY c.id ASC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	defer rows.Close()

	courses := []models.Course{}

	for rows.Next() {
		var course models.Course
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&course.ID,
			&course.Name,
			&course.Description,
			&course.Threshold,
			&createdAtStr,
			&updatedAtStr,
			&course.StudentCount,
			&course.RemainingHours,
		)
		if err != nil {
			return models.PaginatedResponse{}, err
		}

		course.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		course.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		courses = append(courses, course)
	}
	
	totalPages := (total + pageSize - 1) / pageSize

	return models.PaginatedResponse{
		Data:       courses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *CourseService) EnrollStudent(req models.EnrollRequest) error {
	timestamp := db.GetTimestamp()

	query := `INSERT OR IGNORE INTO student_course (student_id, course_id, enrolled_at) VALUES (?, ?, ?)`
	_, err := db.DB.Exec(query, req.StudentID, req.CourseID, timestamp)
	if err != nil {
		return err
	}
	logService := NewOperationLogService()
	logService.Log("create", "course", req.CourseID, fmt.Sprintf("学生选课: %s → %s", resolveStudentName(req.StudentID), resolveCourseName(req.CourseID)))
	return nil
}

func (s *CourseService) UnenrollStudent(req models.EnrollRequest) error {
	query := `DELETE FROM student_course WHERE student_id=? AND course_id=?`
	_, err := db.DB.Exec(query, req.StudentID, req.CourseID)
	if err != nil {
		return err
	}
	logService := NewOperationLogService()
	logService.Log("delete", "course", req.CourseID, fmt.Sprintf("学生退课: %s → %s", resolveStudentName(req.StudentID), resolveCourseName(req.CourseID)))
	return nil
}

func (s *CourseService) GetCourseStudents(courseID int64) ([]models.Student, error) {
	query := `SELECT s.id, s.name, s.student_id, s.contact, s.total_hours, s.completed_hours, s.is_dropped, s.created_at, s.updated_at
			  FROM students s
			  JOIN student_course sc ON s.id = sc.student_id
			  WHERE sc.course_id = ?`

	rows, err := db.DB.Query(query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentService := NewStudentService()
	return studentService.scanStudents(rows)
}

func (s *CourseService) GetCoursesByStudent(studentID int64) ([]models.Course, error) {
	query := `SELECT c.id, c.name, c.description, c.threshold, c.created_at, c.updated_at
		FROM courses c
		JOIN student_course sc ON c.id = sc.course_id
		WHERE sc.student_id = ?
		ORDER BY c.name`

	rows, err := db.DB.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := []models.Course{}

	for rows.Next() {
		var course models.Course
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&course.ID,
			&course.Name,
			&course.Description,
			&course.Threshold,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, err
		}

		course.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		course.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		courses = append(courses, course)
	}

	return courses, nil
}

func (s *CourseService) AddCourseHours(req models.CourseHoursRequest) (*models.Course, error) {
	query := `SELECT id FROM courses WHERE id=?`
	var courseID int64
	err := db.DB.QueryRow(query, req.CourseID).Scan(&courseID)
	if err != nil {
		return nil, err
	}

	// 先收集所有学生 ID，避免在 rows 开启期间执行 UPDATE 导致 SQLITE_BUSY
	query = `SELECT student_id FROM student_course WHERE course_id=?`
	rows, err := db.DB.Query(query, req.CourseID)
	if err != nil {
		return nil, err
	}

	var studentIDs []int64
	for rows.Next() {
		var studentID int64
		if err := rows.Scan(&studentID); err != nil {
			rows.Close()
			return nil, err
		}
		studentIDs = append(studentIDs, studentID)
	}
	rows.Close()

	timestamp := db.GetTimestamp()

	for _, studentID := range studentIDs {
		_, err := db.DB.Exec(`UPDATE students SET total_hours = total_hours + ?, updated_at = ? WHERE id = ?`,
			req.Hours, timestamp, studentID)
		if err != nil {
			return nil, err
		}
	}

	utils.LogInfof("课程添加课时: 课程ID=%d, 课时=%.1f, 学生数=%d", req.CourseID, req.Hours, len(studentIDs))
	logService := NewOperationLogService()
	courseName := resolveCourseName(req.CourseID)
	logService.Log("update", "course", req.CourseID, fmt.Sprintf("批量添加课时: %s (添加 %.1f 课时, 影响 %d 名学生)", courseName, req.Hours, len(studentIDs)))
	return s.GetCourseByID(req.CourseID)
}
