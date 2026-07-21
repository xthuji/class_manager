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

	// 自动关联学生和课程（如果尚未关联）
	_, err = tx.Exec(`INSERT OR IGNORE INTO student_course (student_id, course_id, enrolled_at) VALUES (?, ?, ?)`, req.StudentID, req.CourseID, timestamp)
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

	utils.LogInfof("课时记录创建: ID=%d, 学生ID=%d, 课程ID=%d, 课时=%.1f", id, req.StudentID, req.CourseID, req.Hours)
	logService := NewOperationLogService()
	logService.LogChange("create", "hour_record", id, fmt.Sprintf("%s - %s", resolveStudentName(req.StudentID), resolveCourseName(req.CourseID)), nil)
	return s.GetHourRecordByID(id)
}

func (s *HourRecordService) BatchCreateHourRecord(req models.BatchHourRecordRequest) ([]models.HourRecord, error) {
	timestamp := db.GetTimestamp()
	records := []models.HourRecord{}

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

		// 自动关联学生和课程（如果尚未关联）
		_, err = tx.Exec(`INSERT OR IGNORE INTO student_course (student_id, course_id, enrolled_at) VALUES (?, ?, ?)`, studentID, req.CourseID, timestamp)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		id, _ := result.LastInsertId()
		records = append(records, models.HourRecord{
			ID:          id,
			StudentID:   studentID,
			CourseID:    req.CourseID,
			Hours:       req.Hours,
			RecordDate:  req.RecordDate,
			Description: req.Description,
			CreatedAt:   time.Now(),
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	utils.LogInfof("课时记录批量创建: %d 条, 课程ID=%d, 课时=%.1f", len(records), req.CourseID, req.Hours)
	logService := NewOperationLogService()
	logService.LogChange("create", "hour_record", req.CourseID, fmt.Sprintf("%d条, 课程:[%s %.1f课时], 学生:[%s]", len(records), resolveCourseName(req.CourseID), req.Hours, resolveStudentNames(req.StudentIDs)), nil)
	return records, nil
}

func (s *HourRecordService) GetHourRecordByID(id int64) (*models.HourRecord, error) {
	query := `SELECT id, student_id, course_id, hours, record_date, description, created_at FROM hour_records WHERE id=?`

	row := db.DB.QueryRow(query, id)
	return s.scanHourRecord(row)
}

func (s *HourRecordService) UpdateHourRecord(req models.HourRecordUpdateRequest) (*models.HourRecord, error) {
	// 获取旧记录以计算课时差值
	old, err := s.GetHourRecordByID(req.ID)
	if err != nil {
		return nil, err
	}
	if old == nil {
		return nil, fmt.Errorf("hour record not found")
	}

	timestamp := db.GetTimestamp()
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	// 更新记录
	query := `UPDATE hour_records SET student_id=?, course_id=?, hours=?, record_date=?, description=? WHERE id=?`
	if _, err := tx.Exec(query, req.StudentID, req.CourseID, req.Hours, req.RecordDate, req.Description, req.ID); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 调整旧学生的已完成课时（减去旧课时）
	if old.StudentID != req.StudentID {
		// 学生变了：旧学生减旧课时，新学生加新课时
		if _, err := tx.Exec(`UPDATE students SET completed_hours = completed_hours - ?, updated_at = ? WHERE id = ?`, old.Hours, timestamp, old.StudentID); err != nil {
			tx.Rollback()
			return nil, err
		}
		if _, err := tx.Exec(`UPDATE students SET completed_hours = completed_hours + ?, updated_at = ? WHERE id = ?`, req.Hours, timestamp, req.StudentID); err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		// 学生不变：调整课时差值
		diff := req.Hours - old.Hours
		if _, err := tx.Exec(`UPDATE students SET completed_hours = completed_hours + ?, updated_at = ? WHERE id = ?`, diff, timestamp, req.StudentID); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	utils.LogInfof("课时记录更新: ID=%d, 学生ID=%d, 课程ID=%d, 课时=%.1f->%.1f", req.ID, req.StudentID, req.CourseID, old.Hours, req.Hours)
	logService := NewOperationLogService()
	var changes []FieldChange
	if old.StudentID != req.StudentID {
		changes = append(changes, FieldChange{Field: "学生", Old: resolveStudentName(old.StudentID), New: resolveStudentName(req.StudentID)})
	}
	if old.CourseID != req.CourseID {
		changes = append(changes, FieldChange{Field: "课程", Old: resolveCourseName(old.CourseID), New: resolveCourseName(req.CourseID)})
	}
	if old.RecordDate != req.RecordDate {
		changes = append(changes, FieldChange{Field: "日期", Old: old.RecordDate, New: req.RecordDate})
	}
	if old.Hours != req.Hours {
		changes = append(changes, FieldChange{Field: "课时", Old: old.Hours, New: req.Hours})
	}
	if old.Description != req.Description {
		changes = append(changes, FieldChange{Field: "备注", Old: old.Description, New: req.Description})
	}
	logService.LogChange("update", "hour_record", req.ID, resolveStudentName(req.StudentID), changes)
	return s.GetHourRecordByID(req.ID)
}

func (s *HourRecordService) DeleteHourRecord(id int64) error {
	// 获取记录以知道要扣减的课时
	old, err := s.GetHourRecordByID(id)
	if err != nil {
		return err
	}
	if old == nil {
		return fmt.Errorf("hour record not found")
	}

	timestamp := db.GetTimestamp()
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM hour_records WHERE id=?`, id); err != nil {
		tx.Rollback()
		return err
	}

	// 扣减学生的已完成课时
	if _, err := tx.Exec(`UPDATE students SET completed_hours = completed_hours - ?, updated_at = ? WHERE id = ?`, old.Hours, timestamp, old.StudentID); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	utils.LogInfof("课时记录删除: ID=%d, 学生ID=%d, 课时=%.1f", id, old.StudentID, old.Hours)
	logService := NewOperationLogService()
	logService.LogChange("delete", "hour_record", id, fmt.Sprintf("%s - %s", resolveStudentName(old.StudentID), resolveCourseName(old.CourseID)), nil)
	return nil
}

// BatchDeleteHourRecords 批量删除课时记录（每条删除都会返还学生课时）
func (s *HourRecordService) BatchDeleteHourRecords(ids []int64) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.DeleteHourRecord(id); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
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

func (s *HourRecordService) ListHourRecords(req models.HourRecordListRequest) (models.PaginatedResponse, error) {
	query := `SELECT hr.id, hr.student_id, hr.course_id, hr.hours, hr.record_date, hr.description, hr.created_at,
			  s.name as student_name, s.student_id as student_no, c.name as course_name
			  FROM hour_records hr 
			  JOIN students s ON hr.student_id = s.id 
			  LEFT JOIN courses c ON hr.course_id = c.id 
			  WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM hour_records hr 
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
		countQuery += " AND hr.student_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	if len(req.CourseIDs) > 0 {
		placeholders := make([]string, len(req.CourseIDs))
		for i, id := range req.CourseIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += " AND hr.course_id IN (" + strings.Join(placeholders, ",") + ")"
		countQuery += " AND hr.course_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	if req.StartDate != "" {
		query += " AND hr.record_date >= ?"
		countQuery += " AND hr.record_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND hr.record_date <= ?"
		countQuery += " AND hr.record_date <= ?"
		args = append(args, req.EndDate)
	}

	// 学生状态筛选
	if req.StudentStatus != "" {
		query += " AND s.is_dropped = ?"
		countQuery += " AND s.is_dropped = ?"
		args = append(args, req.StudentStatus)
	}

	// 获取总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	var total int
	err := db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	query += " ORDER BY hr.record_date DESC, hr.id DESC"

	// 分页
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	query += " LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return models.PaginatedResponse{}, err
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
		)

		if err != nil {
			return models.PaginatedResponse{}, err
		}

		record.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		records = append(records, record)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return models.PaginatedResponse{
		Data:       records,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *HourRecordService) GetDistinctDates() ([]string, error) {
	query := `SELECT DISTINCT record_date FROM hour_records ORDER BY record_date DESC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := []string{}
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

	if len(req.StudentIDs) > 0 {
		placeholders := make([]string, len(req.StudentIDs))
		for i := range req.StudentIDs {
			placeholders[i] = "?"
			args = append(args, req.StudentIDs[i])
		}
		query += " AND hr.student_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	if len(req.CourseIDs) > 0 {
		placeholders := make([]string, len(req.CourseIDs))
		for i := range req.CourseIDs {
			placeholders[i] = "?"
			args = append(args, req.CourseIDs[i])
		}
		query += " AND hr.course_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	if req.StudentStatus != "" {
		query += " AND s.status = ?"
		args = append(args, req.StudentStatus)
	}

	if req.StartDate != "" {
		query += " AND hr.record_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND hr.record_date <= ?"
		args = append(args, req.EndDate)
	}

	query += " ORDER BY hr.record_date DESC, hr.id DESC"

	if req.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, req.Limit)
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("日期,课时,备注,学生姓名,学号,课程名称\n")

	for rows.Next() {
		var recordDate, description, studentName, studentNo, courseName string
		var hours float64

		if err := rows.Scan(&recordDate, &hours, &description, &studentName, &studentNo, &courseName); err != nil {
			return "", err
		}

		sb.WriteString(fmt.Sprintf("%s,%.2f,%s,%s,%s,%s\n",
			recordDate, hours, description, studentName, studentNo, courseName))
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
	records := []models.HourRecord{}

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