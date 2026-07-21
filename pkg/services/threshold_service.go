package services

import (
	"database/sql"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
)

type ThresholdService struct{}

func NewThresholdService() *ThresholdService {
	return &ThresholdService{}
}

func (s *ThresholdService) SetThreshold(req models.ThresholdRequest) (*models.Threshold, error) {
	timestamp := db.GetTimestamp()

	query := `INSERT OR REPLACE INTO threshold (subject, value, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err := db.DB.Exec(query, req.Subject, req.Value, timestamp, timestamp)
	if err != nil {
		return nil, err
	}

	return s.GetThreshold(req.Subject)
}

func (s *ThresholdService) GetThreshold(subject string) (*models.Threshold, error) {
	query := `SELECT id, subject, value, template_name, created_at, updated_at FROM threshold WHERE subject=?`

	row := db.DB.QueryRow(query, subject)
	return s.scanThreshold(row)
}

func (s *ThresholdService) ListThresholds() ([]models.Threshold, error) {
	query := `SELECT id, subject, value, template_name, created_at, updated_at FROM threshold`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanThresholds(rows)
}

func (s *ThresholdService) SaveThresholdTemplate(req models.ThresholdTemplateRequest) error {
	timestamp := db.GetTimestamp()

	for _, threshold := range req.Thresholds {
		query := `INSERT OR REPLACE INTO threshold (subject, value, template_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
		_, err := db.DB.Exec(query, threshold.Subject, threshold.Value, req.TemplateName, timestamp, timestamp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ThresholdService) ApplyThresholdTemplate(templateName string) error {
	query := `SELECT subject, value FROM threshold WHERE template_name=?`

	rows, err := db.DB.Query(query, templateName)
	if err != nil {
		return err
	}
	defer rows.Close()

	timestamp := db.GetTimestamp()

	for rows.Next() {
		var subject string
		var value float64

		if err := rows.Scan(&subject, &value); err != nil {
			return err
		}

		updateQuery := `UPDATE threshold SET value=?, updated_at=? WHERE subject=?`
		_, err := db.DB.Exec(updateQuery, value, timestamp, subject)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ThresholdService) PreviewThreshold(subject string, value float64) (models.ThresholdPreviewResponse, error) {
	var preview models.ThresholdPreviewResponse
	preview.Subject = subject
	preview.ThresholdValue = value

	query := `SELECT COUNT(*) FROM students WHERE subjects LIKE ?`
	err := db.DB.QueryRow(query, "%"+subject+"%").Scan(&preview.TotalStudentCount)
	if err != nil {
		return preview, err
	}

	query = `SELECT COUNT(*) FROM students WHERE subjects LIKE ? AND (total_hours - completed_hours) < ?`
	err = db.DB.QueryRow(query, "%"+subject+"%", value).Scan(&preview.TriggerCount)
	if err != nil {
		return preview, err
	}

	return preview, nil
}

func (s *ThresholdService) scanThreshold(row *sql.Row) (*models.Threshold, error) {
	var threshold models.Threshold
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&threshold.ID,
		&threshold.Subject,
		&threshold.Value,
		&threshold.TemplateName,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	threshold.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	threshold.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &threshold, nil
}

func (s *ThresholdService) scanThresholds(rows *sql.Rows) ([]models.Threshold, error) {
	var thresholds []models.Threshold

	for rows.Next() {
		var threshold models.Threshold
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&threshold.ID,
			&threshold.Subject,
			&threshold.Value,
			&threshold.TemplateName,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		threshold.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		threshold.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		thresholds = append(thresholds, threshold)
	}

	return thresholds, nil
}
