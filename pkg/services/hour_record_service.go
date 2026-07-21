package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
)

type HourRecordService struct{}

func NewHourRecordService() *HourRecordService {
	return &HourRecordService{}
}

func (s *HourRecordService) CreateHourRecord(req models.HourRecordCreateRequest) (*models.HourRecord, error) {
	timestamp := db.GetTimestamp()

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO hour_records (student_id, course_id, hours, record_date, description, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(query, req.StudentID, req.CourseID, req.Hours, req.RecordDate, req.Description, timestamp)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	updateQuery := `UPDATE students SET completed_hours = completed_hours + ?, updated_at = ? WHERE id = ?`
	_, err = tx.Exec(updateQuery, req.Hours, timestamp, req.StudentID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetHourRecordByID(id)
}

func (s *HourRecordService) BatchCreateHourRecord(req models.BatchHourRecordRequest) ([]models.HourRecord, error) {
	timestamp := db.GetTimestamp()
	var records []models.HourRecord

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	for _, studentID := range req.StudentIDs {
		query := `INSERT INTO hour_records (student_id, course_id, hours, record_date, description, created_at) 
				  VALUES (?, ?, ?, ?, ?, ?)`

		result, err := tx.Exec(query, studentID, req.CourseID, req.Hours, req.RecordDate, req.Description, timestamp)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		updateQuery := `UPDATE students SET completed_hours = completed_hours + ?, updated_at = ? WHERE id = ?`
		_, err = tx.Exec(updateQuery, req.Hours, timestamp, studentID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		id, _ := result.LastInsertId()
		record, _ := s.GetHourRecordByID(id)
		if record != nil {
			records = append(records, *record)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *HourRecordService) GetHourRecordByID(id int64) (*models.HourRecord, error) {
	query := `SELECT id, student_id, course_id, hours, record_date, description, created_at FROM hour_records WHERE id=?`

	row := db.DB.QueryRow(query, id)
	return s.scanHourRecord(row)
}

func (s *HourRecordService) GetHourRecordsByStudent(studentID int64) ([]models.HourRecord, error) {
	query := `SELECT id, student_id, course_id, hours, record_date, description, created_at FROM hour_records WHERE student_id=? ORDER BY record_date DESC`

	rows, err := db.DB.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanHourRecords(rows)
}

func (s *HourRecordService) ListHourRecords(req models.HourRecordListRequest) ([]models.HourRecordWithDetails, error) {
	query := `SELECT hr.id, hr.student_id, hr.course_id, hr.hours, hr.record_date, hr.description, hr.created_at,
			  s.name as student_name, s.student_id as student_no, c.name as course_name, c.subject as course_subject
			  FROM hour_records hr 
			  JOIN students s ON hr.student_id = s.id 
			  LEFT JOIN courses c ON hr.course_id = c.id 
			  WHERE 1=1`

	var args []interface{}

	if len(req.StudentIDs) > 0 {
		placeholders := make([]string, len(req.StudentIDs))
		for i, id := range req.StudentIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += " AND hr.student_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	if len(req.CourseIDs) > 0 {
		placeholders := make([]string, len(req.CourseIDs))
		for i, id := range req.CourseIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += " AND hr.course_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	if req.StartDate != "" {
		query += " AND hr.record_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND hr.record_date <= ?"
		args = append(args, req.EndDate)
	}

	query += " ORDER BY hr.record_date DESC, hr.created_at DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.HourRecordWithDetails
	for rows.Next() {
		var record models.HourRecordWithDetails
		var createdAtStr string

		err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.CourseID,
			&record.Hours,
			&record.RecordDate,
			&record.Description,
			&createdAtStr,
			&record.StudentName,
			&record.StudentNo,
			&record.CourseName,
			&record.CourseSubject,
		)

		if err != nil {
			return nil, err
		}

		record.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		records = append(records, record)
	}

	return records, nil
}

func (s *HourRecordService) GetDistinctDates() ([]string, error) {
	query := `SELECT DISTINCT record_date FROM hour_records ORDER BY record_date DESC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}

	return dates, nil
}

func (s *HourRecordService) ExportHourRecords(req models.ExportRequest) (string, error) {
	query := `SELECT hr.record_date, hr.hours, hr.description, s.name, s.student_id, c.name as course_name 
			  FROM hour_records hr 
			  JOIN students s ON hr.student_id = s.id 
			  LEFT JOIN courses c ON hr.course_id = c.id 
			  WHERE 1=1`

	var args []interface{}

	if req.StudentID != 0 {
		query += " AND hr.student_id = ?"
		args = append(args, req.StudentID)
	}

	if req.StartDate != "" {
		query += " AND hr.record_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND hr.record_date <= ?"
		args = append(args, req.EndDate)
	}

	query += " ORDER BY hr.record_date DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("日期,课时,备注,学生姓名,学号,课程名称\n")

	for rows.Next() {
		var recordDate, description, studentName, studentID, courseName string
		var hours float64

		if err := rows.Scan(&recordDate, &hours, &description, &studentName, &studentID, &courseName); err != nil {
			return "", err
		}

		sb.WriteString(fmt.Sprintf("%s,%.2f,%s,%s,%s,%s\n",
			recordDate, hours, description, studentName, studentID, courseName))
	}

	return sb.String(), nil
}

func (s *HourRecordService) GetStudentHoursSummary(studentID int64) (models.HoursSummaryResponse, error) {
	var summary models.HoursSummaryResponse

	query := `SELECT total_hours, completed_hours FROM students WHERE id=?`
	err := db.DB.QueryRow(query, studentID).Scan(&summary.TotalHours, &summary.CompletedHours)
	if err != nil {
		return summary, err
	}

	summary.RemainingHours = summary.TotalHours - summary.CompletedHours
	return summary, nil
}

func (s *HourRecordService) scanHourRecord(row *sql.Row) (*models.HourRecord, error) {
	var record models.HourRecord
	var createdAtStr string

	err := row.Scan(
		&record.ID,
		&record.StudentID,
		&record.CourseID,
		&record.Hours,
		&record.RecordDate,
		&record.Description,
		&createdAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	record.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	return &record, nil
}

func (s *HourRecordService) scanHourRecords(rows *sql.Rows) ([]models.HourRecord, error) {
	var records []models.HourRecord

	for rows.Next() {
		var record models.HourRecord
		var createdAtStr string

		err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.CourseID,
			&record.Hours,
			&record.RecordDate,
			&record.Description,
			&createdAtStr,
		)

		if err != nil {
			return nil, err
		}

		record.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		records = append(records, record)
	}

	return records, nil
}