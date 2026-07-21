package services

import (
	"database/sql"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
)

type CourseService struct{}

func NewCourseService() *CourseService {
	return &CourseService{}
}

func (s *CourseService) CreateCourse(req models.CourseCreateRequest) (*models.Course, error) {
	timestamp := db.GetTimestamp()

	query := `INSERT INTO courses (name, subject, description, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?)`

	result, err := db.DB.Exec(query, req.Name, req.Subject, req.Description, timestamp, timestamp)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetCourseByID(id)
}

func (s *CourseService) UpdateCourse(req models.CourseUpdateRequest) (*models.Course, error) {
	timestamp := db.GetTimestamp()

	query := `UPDATE courses SET name=?, subject=?, description=?, updated_at=? WHERE id=?`

	_, err := db.DB.Exec(query, req.Name, req.Subject, req.Description, timestamp, req.ID)
	if err != nil {
		return nil, err
	}

	return s.GetCourseByID(req.ID)
}

func (s *CourseService) DeleteCourse(id int64) error {
	query := `DELETE FROM courses WHERE id=?`
	_, err := db.DB.Exec(query, id)
	return err
}

func (s *CourseService) GetCourseByID(id int64) (*models.Course, error) {
	query := `SELECT id, name, subject, description, created_at, updated_at FROM courses WHERE id=?`

	row := db.DB.QueryRow(query, id)
	return s.scanCourse(row)
}

func (s *CourseService) ListCourses(req models.CourseListRequest) ([]models.Course, error) {
	query := `SELECT id, name, subject, description, created_at, updated_at FROM courses`
	var args []interface{}

	if req.Subject != "" {
		query += " WHERE subject=?"
		args = append(args, req.Subject)
	}

	offset := (req.Page - 1) * req.PageSize
	query += " LIMIT ? OFFSET ?"
	args = append(args, req.PageSize, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanCourses(rows)
}

func (s *CourseService) EnrollStudent(req models.EnrollRequest) error {
	timestamp := db.GetTimestamp()

	query := `INSERT OR IGNORE INTO student_course (student_id, course_id, enrolled_at) VALUES (?, ?, ?)`
	_, err := db.DB.Exec(query, req.StudentID, req.CourseID, timestamp)
	return err
}

func (s *CourseService) GetCourseStudents(courseID int64) ([]models.Student, error) {
	query := `SELECT s.id, s.name, s.student_id, s.contact, s.subjects, s.total_hours, s.completed_hours, s.created_at, s.updated_at 
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

func (s *CourseService) AddCourseHours(req models.CourseHoursRequest) (*models.Course, error) {
	query := `SELECT id FROM courses WHERE id=?`
	var courseID int64
	err := db.DB.QueryRow(query, req.CourseID).Scan(&courseID)
	if err != nil {
		return nil, err
	}

	query = `SELECT student_id FROM student_course WHERE course_id=?`
	rows, err := db.DB.Query(query, req.CourseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	timestamp := db.GetTimestamp()

	for rows.Next() {
		var studentID int64
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}

		_, err := db.DB.Exec(`UPDATE students SET total_hours = total_hours + ?, updated_at = ? WHERE id = ?`,
			req.Hours, timestamp, studentID)
		if err != nil {
			return nil, err
		}
	}

	return s.GetCourseByID(req.CourseID)
}

func (s *CourseService) scanCourse(row *sql.Row) (*models.Course, error) {
	var course models.Course
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&course.ID,
		&course.Name,
		&course.Subject,
		&course.Description,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	course.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	course.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	countQuery := `SELECT COUNT(*) FROM student_course WHERE course_id=?`
	db.DB.QueryRow(countQuery, course.ID).Scan(&course.StudentCount)

	return &course, nil
}

func (s *CourseService) scanCourses(rows *sql.Rows) ([]models.Course, error) {
	var courses []models.Course

	for rows.Next() {
		var course models.Course
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&course.ID,
			&course.Name,
			&course.Subject,
			&course.Description,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		course.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		course.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		countQuery := `SELECT COUNT(*) FROM student_course WHERE course_id=?`
		db.DB.QueryRow(countQuery, course.ID).Scan(&course.StudentCount)

		courses = append(courses, course)
	}

	return courses, nil
}