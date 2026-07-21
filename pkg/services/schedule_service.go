package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/utils"
)

type ScheduleService struct{}

func NewScheduleService() *ScheduleService {
	return &ScheduleService{}
}

func (s *ScheduleService) CreateSchedule(req models.ScheduleCreateRequest) (*models.Schedule, error) {
	timestamp := db.GetTimestamp()

	query := `INSERT INTO schedules (course_id, day_of_week, period, start_time, end_time, hours_consumed, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.DB.Exec(query, req.CourseID, req.DayOfWeek, req.Period, req.StartTime, req.EndTime, req.HoursConsumed, timestamp, timestamp)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	utils.LogInfof("排课创建: ID=%d, 课程ID=%d", id, req.CourseID)
	logService := NewOperationLogService()
	logService.LogChange("update", "course", req.CourseID, resolveCourseName(req.CourseID), []FieldChange{{Field: "排课", Old: "", New: fmt.Sprintf("周%d %s", req.DayOfWeek, req.Period)}})
	return s.GetScheduleByID(id)
}

func (s *ScheduleService) UpdateSchedule(req models.ScheduleUpdateRequest) (*models.Schedule, error) {
	old, _ := s.GetScheduleByID(req.ID)

	timestamp := db.GetTimestamp()

	query := `UPDATE schedules SET day_of_week=?, period=?, start_time=?, end_time=?, hours_consumed=?, updated_at=? WHERE id=?`

	_, err := db.DB.Exec(query, req.DayOfWeek, req.Period, req.StartTime, req.EndTime, req.HoursConsumed, timestamp, req.ID)
	if err != nil {
		return nil, err
	}

	utils.LogInfof("排课更新: ID=%d", req.ID)
	logService := NewOperationLogService()
	var changes []FieldChange
	var entityName string
	var courseID int64
	if old != nil {
		courseID = old.CourseID
		entityName = resolveCourseName(old.CourseID)
		if old.DayOfWeek != req.DayOfWeek {
			changes = append(changes, FieldChange{Field: "排课-星期", Old: old.DayOfWeek, New: req.DayOfWeek})
		}
		if old.Period != req.Period {
			changes = append(changes, FieldChange{Field: "排课-时间段", Old: old.Period, New: req.Period})
		}
		if old.StartTime != req.StartTime {
			changes = append(changes, FieldChange{Field: "排课-开始时间", Old: old.StartTime, New: req.StartTime})
		}
		if old.EndTime != req.EndTime {
			changes = append(changes, FieldChange{Field: "排课-结束时间", Old: old.EndTime, New: req.EndTime})
		}
		if old.HoursConsumed != req.HoursConsumed {
			changes = append(changes, FieldChange{Field: "排课-消耗课时", Old: old.HoursConsumed, New: req.HoursConsumed})
		}
	}
	if len(changes) > 0 {
		logService.LogChange("update", "course", courseID, entityName, changes)
	}
	return s.GetScheduleByID(req.ID)
}

func (s *ScheduleService) DeleteSchedule(id int64) error {
	old, _ := s.GetScheduleByID(id)

	query := `DELETE FROM schedules WHERE id=?`
	_, err := db.DB.Exec(query, id)
	if err != nil {
		return err
	}
	utils.LogInfof("排课删除: ID=%d", id)
	logService := NewOperationLogService()
	if old != nil {
		logService.LogChange("update", "course", old.CourseID, resolveCourseName(old.CourseID), []FieldChange{{Field: "排课", Old: fmt.Sprintf("周%d %s", old.DayOfWeek, old.Period), New: ""}})
	}
	return nil
}

func (s *ScheduleService) GetScheduleByID(id int64) (*models.Schedule, error) {
	query := `SELECT id, course_id, day_of_week, period, start_time, end_time, hours_consumed, created_at, updated_at FROM schedules WHERE id=?`

	row := db.DB.QueryRow(query, id)
	return s.scanSchedule(row)
}

func (s *ScheduleService) GetCourseSchedule(courseID int64) ([]models.Schedule, error) {
	query := `SELECT id, course_id, day_of_week, period, start_time, end_time, hours_consumed, created_at, updated_at FROM schedules WHERE course_id=? ORDER BY day_of_week, start_time`

	rows, err := db.DB.Query(query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanSchedules(rows)
}

func (s *ScheduleService) scanSchedule(row *sql.Row) (*models.Schedule, error) {
	var schedule models.Schedule
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&schedule.ID,
		&schedule.CourseID,
		&schedule.DayOfWeek,
		&schedule.Period,
		&schedule.StartTime,
		&schedule.EndTime,
		&schedule.HoursConsumed,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	schedule.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	schedule.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &schedule, nil
}

func (s *ScheduleService) scanSchedules(rows *sql.Rows) ([]models.Schedule, error) {
	schedules := []models.Schedule{}

	for rows.Next() {
		var schedule models.Schedule
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&schedule.ID,
			&schedule.CourseID,
			&schedule.DayOfWeek,
			&schedule.Period,
			&schedule.StartTime,
			&schedule.EndTime,
			&schedule.HoursConsumed,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		schedule.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		schedule.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}
