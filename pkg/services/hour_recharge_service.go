package services

import (
	"database/sql"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
)

type HourRechargeService struct{}

func NewHourRechargeService() *HourRechargeService {
	return &HourRechargeService{}
}

func (s *HourRechargeService) CreateRecharge(req models.HourRechargeCreateRequest) (*models.HourRecharge, error) {
	timestamp := db.GetTimestamp()

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `INSERT INTO hour_recharges (student_id, hours, recharge_date, description, created_at) 
			  VALUES (?, ?, ?, ?, ?)`

	result, err := tx.Exec(query, req.StudentID, req.Hours, req.RechargeDate, req.Description, timestamp)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`UPDATE students SET total_hours = total_hours + ?, updated_at = ? WHERE id = ?`,
		req.Hours, timestamp, req.StudentID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetRechargeByID(id)
}

func (s *HourRechargeService) GetRechargeByID(id int64) (*models.HourRecharge, error) {
	query := `SELECT hr.id, hr.student_id, s.name, s.student_id, hr.hours, hr.recharge_date, hr.description, hr.created_at 
			  FROM hour_recharges hr 
			  JOIN students s ON hr.student_id = s.id 
			  WHERE hr.id = ?`

	row := db.DB.QueryRow(query, id)
	return s.scanRecharge(row)
}

func (s *HourRechargeService) ListRecharges(req models.HourRechargeListRequest) ([]models.HourRecharge, error) {
	query := `SELECT hr.id, hr.student_id, s.name, s.student_id, hr.hours, hr.recharge_date, hr.description, hr.created_at 
			  FROM hour_recharges hr 
			  JOIN students s ON hr.student_id = s.id 
			  WHERE 1=1`

	var args []interface{}

	if req.StudentID != 0 {
		query += " AND hr.student_id = ?"
		args = append(args, req.StudentID)
	}

	if req.StartDate != "" {
		query += " AND hr.recharge_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND hr.recharge_date <= ?"
		args = append(args, req.EndDate)
	}

	query += " ORDER BY hr.recharge_date DESC"

	offset := (req.Page - 1) * req.PageSize
	query += " LIMIT ? OFFSET ?"
	args = append(args, req.PageSize, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanRecharges(rows)
}

func (s *HourRechargeService) DeleteRecharge(id int64) error {
	query := `DELETE FROM hour_recharges WHERE id=?`
	_, err := db.DB.Exec(query, id)
	return err
}

func (s *HourRechargeService) scanRecharge(row *sql.Row) (*models.HourRecharge, error) {
	var recharge models.HourRecharge
	var createdAtStr string

	err := row.Scan(
		&recharge.ID,
		&recharge.StudentID,
		&recharge.StudentName,
		&recharge.StudentNo,
		&recharge.Hours,
		&recharge.RechargeDate,
		&recharge.Description,
		&createdAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	recharge.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

	return &recharge, nil
}

func (s *HourRechargeService) scanRecharges(rows *sql.Rows) ([]models.HourRecharge, error) {
	var recharges []models.HourRecharge

	for rows.Next() {
		var recharge models.HourRecharge
		var createdAtStr string

		err := rows.Scan(
			&recharge.ID,
			&recharge.StudentID,
			&recharge.StudentName,
			&recharge.StudentNo,
			&recharge.Hours,
			&recharge.RechargeDate,
			&recharge.Description,
			&createdAtStr,
		)

		if err != nil {
			return nil, err
		}

		recharge.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		recharges = append(recharges, recharge)
	}

	return recharges, nil
}