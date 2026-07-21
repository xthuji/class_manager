package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/utils"
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

	utils.LogInfof("充值创建: 学生ID=%d, 课时=%.1f", req.StudentID, req.Hours)
	recharge, err := s.GetRechargeByID(id)
	if err != nil {
		return nil, err
	}
	logService := NewOperationLogService()
	logService.LogChange("create", "hour_recharge", id, recharge.StudentName, nil)
	return recharge, nil
}

func (s *HourRechargeService) GetRechargeByID(id int64) (*models.HourRecharge, error) {
	query := `SELECT hr.id, hr.student_id, s.name, s.student_id, hr.hours, hr.recharge_date, hr.description, hr.created_at 
			  FROM hour_recharges hr 
			  JOIN students s ON hr.student_id = s.id 
			  WHERE hr.id = ?`

	row := db.DB.QueryRow(query, id)
	return s.scanRecharge(row)
}

func (s *HourRechargeService) ListRecharges(req models.HourRechargeListRequest) (models.PaginatedResponse, error) {
	query := `SELECT hr.id, hr.student_id, s.name, s.student_id, hr.hours, hr.recharge_date, hr.description, hr.created_at 
			  FROM hour_recharges hr 
			  JOIN students s ON hr.student_id = s.id 
			  WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM hour_recharges hr 
			  JOIN students s ON hr.student_id = s.id 
			  WHERE 1=1`

	var args []interface{}

	if req.StudentID != 0 {
		query += " AND hr.student_id = ?"
		countQuery += " AND hr.student_id = ?"
		args = append(args, req.StudentID)
	}

	if req.StartDate != "" {
		query += " AND hr.recharge_date >= ?"
		countQuery += " AND hr.recharge_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND hr.recharge_date <= ?"
		countQuery += " AND hr.recharge_date <= ?"
		args = append(args, req.EndDate)
	}

	// 获取总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	var total int
	err := db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	query += " ORDER BY hr.id DESC"

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

	recharges, err := s.scanRecharges(rows)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	
	totalPages := (total + pageSize - 1) / pageSize

	return models.PaginatedResponse{
		Data:       recharges,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *HourRechargeService) GetRechargesByStudent(studentID int64) ([]models.HourRecharge, error) {
	result, err := s.ListRecharges(models.HourRechargeListRequest{
		StudentID: studentID,
		Page:      1,
		PageSize:  100,
	})
	if err != nil {
		return nil, err
	}
	return result.Data.([]models.HourRecharge), nil
}

func (s *HourRechargeService) DeleteRecharge(id int64) error {
	old, err := s.GetRechargeByID(id)
	if err != nil {
		return err
	}
	if old == nil {
		return fmt.Errorf("recharge not found")
	}

	timestamp := db.GetTimestamp()
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM hour_recharges WHERE id=?`, id); err != nil {
		tx.Rollback()
		return err
	}

	// 扣减学生的总课时（充值时增加了总课时，删除时需返还）
	if _, err := tx.Exec(`UPDATE students SET total_hours = total_hours - ?, updated_at = ? WHERE id = ?`, old.Hours, timestamp, old.StudentID); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	utils.LogInfof("充值删除: ID=%d, 学生ID=%d, 课时=%.1f", id, old.StudentID, old.Hours)
	logService := NewOperationLogService()
	logService.LogChange("delete", "hour_recharge", id, old.StudentName, nil)
	return nil
}

// BatchDeleteRecharges 批量删除充值记录（每条删除都会扣减学生课时）
func (s *HourRechargeService) BatchDeleteRecharges(ids []int64) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.DeleteRecharge(id); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *HourRechargeService) UpdateRecharge(req models.HourRechargeUpdateRequest) (*models.HourRecharge, error) {
	// Get old recharge to calculate hours difference
	old, err := s.GetRechargeByID(req.ID)
	if err != nil {
		return nil, err
	}
	if old == nil {
		return nil, fmt.Errorf("recharge not found")
	}

	timestamp := db.GetTimestamp()
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update recharge record
	_, err = tx.Exec(`UPDATE hour_recharges SET hours=?, recharge_date=?, description=? WHERE id=?`,
		req.Hours, req.RechargeDate, req.Description, req.ID)
	if err != nil {
		return nil, err
	}

	// Adjust student total_hours: subtract old hours, add new hours
	hoursDiff := req.Hours - old.Hours
	if hoursDiff != 0 {
		_, err = tx.Exec(`UPDATE students SET total_hours = total_hours + ?, updated_at = ? WHERE id = ?`,
			hoursDiff, timestamp, old.StudentID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	utils.LogInfof("充值更新: ID=%d, 课时=%.1f→%.1f", req.ID, old.Hours, req.Hours)
	logService := NewOperationLogService()
	var changes []FieldChange
	if old.Hours != req.Hours {
		changes = append(changes, FieldChange{Field: "课时", Old: old.Hours, New: req.Hours})
	}
	if old.RechargeDate != req.RechargeDate {
		changes = append(changes, FieldChange{Field: "充值日期", Old: old.RechargeDate, New: req.RechargeDate})
	}
	if old.Description != req.Description {
		changes = append(changes, FieldChange{Field: "备注", Old: old.Description, New: req.Description})
	}
	logService.LogChange("update", "hour_recharge", req.ID, old.StudentName, changes)
	return s.GetRechargeByID(req.ID)
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
	recharges := []models.HourRecharge{}

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