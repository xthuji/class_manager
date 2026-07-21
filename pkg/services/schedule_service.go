package services

import (
	"database/sql"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
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

	return s.GetScheduleByID(id)
}

func (s *ScheduleService) UpdateSchedule(req models.ScheduleUpdateRequest) (*models.Schedule, error) {
	timestamp := db.GetTimestamp()

	query := `UPDATE schedules SET day_of_week=?, period=?, start_time=?, end_time=?, hours_consumed=?, updated_at=? WHERE id=?`

	_, err := db.DB.Exec(query, req.DayOfWeek, req.Period, req.StartTime, req.EndTime, req.HoursConsumed, timestamp, req.ID)
	if err != nil {
		return nil, err
	}

	return s.GetScheduleByID(req.ID)
}

func (s *ScheduleService) DeleteSchedule(id int64) error {
	query := `DELETE FROM schedules WHERE id=?`
	_, err := db.DB.Exec(query, id)
	return err
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
	var schedules []models.Schedule

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
