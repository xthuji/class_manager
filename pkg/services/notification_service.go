package services

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) GetNotifications(status string) ([]models.Notification, error) {
	query := `SELECT n.id, n.student_id, n.subject, n.current_hours, n.threshold, n.status, n.created_at, n.processed_at, s.name as student_name 
			  FROM notifications n 
			  JOIN students s ON n.student_id = s.id`

	if status != "" {
		query += " WHERE n.status=?"
	}

	rows, err := db.DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanNotifications(rows)
}

func (s *NotificationService) GetUnreadCount() (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE status='unread'`

	var count int
	err := db.DB.QueryRow(query).Scan(&count)
	return count, err
}

func (s *NotificationService) MarkAsRead(id int64) error {
	query := `UPDATE notifications SET status='read' WHERE id=?`
	_, err := db.DB.Exec(query, id)
	return err
}

func (s *NotificationService) BatchMarkAsRead(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))

	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `UPDATE notifications SET status='read' WHERE id IN (` + strings.Join(placeholders, ",") + `)`
	_, err := db.DB.Exec(query, args...)
	return err
}

func (s *NotificationService) MarkAsProcessed(id int64) error {
	timestamp := db.GetTimestamp()
	query := `UPDATE notifications SET status='processed', processed_at=? WHERE id=?`
	_, err := db.DB.Exec(query, timestamp, id)
	return err
}

func (s *NotificationService) BatchMarkAsProcessed(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	timestamp := db.GetTimestamp()
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)

	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	args[len(ids)] = timestamp

	query := `UPDATE notifications SET status='processed', processed_at=? WHERE id IN (` + strings.Join(placeholders, ",") + `)`
	_, err := db.DB.Exec(query, args...)
	return err
}

func (s *NotificationService) CreateNotification(studentID int64, subject string, currentHours, threshold float64) error {
	timestamp := db.GetTimestamp()

	query := `INSERT INTO notifications (student_id, subject, current_hours, threshold, status, created_at) VALUES (?, ?, ?, ?, 'unread', ?)`
	_, err := db.DB.Exec(query, studentID, subject, currentHours, threshold, timestamp)
	if err != nil {
		return err
	}

	return s.SendSystemNotification(models.NotificationRequest{
		Title:    "课时阈值提醒",
		Subtitle: subject,
		Body:     fmt.Sprintf("学生课时已低于阈值，当前剩余%.2f课时，阈值为%.2f课时", currentHours, threshold),
	})
}

func (s *NotificationService) SendSystemNotification(req models.NotificationRequest) error {
	script := `display notification "` + escapeAppleScript(req.Body) + `" with title "` + escapeAppleScript(req.Title) + `" subtitle "` + escapeAppleScript(req.Subtitle) + `"`
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func (s *NotificationService) CheckThresholds() error {
	thresholdService := NewThresholdService()
	thresholds, err := thresholdService.ListThresholds()
	if err != nil {
		return err
	}

	for _, threshold := range thresholds {
		query := `SELECT id, (total_hours - completed_hours) as remaining_hours FROM students WHERE subjects LIKE ? AND (total_hours - completed_hours) < ?`
		rows, err := db.DB.Query(query, "%"+threshold.Subject+"%", threshold.Value)
		if err != nil {
			return err
		}

		for rows.Next() {
			var studentID int64
			var remainingHours float64

			if err := rows.Scan(&studentID, &remainingHours); err != nil {
				rows.Close()
				return err
			}

			checkQuery := `SELECT COUNT(*) FROM notifications WHERE student_id=? AND subject=? AND status IN ('unread', 'read')`
			var exists int
			db.DB.QueryRow(checkQuery, studentID, threshold.Subject).Scan(&exists)

			if exists == 0 {
				s.CreateNotification(studentID, threshold.Subject, remainingHours, threshold.Value)
			}
		}

		rows.Close()
	}

	return nil
}

func escapeAppleScript(str string) string {
	str = strings.ReplaceAll(str, `"`, `\"`)
	str = strings.ReplaceAll(str, `\`, `\\`)
	return str
}

func (s *NotificationService) scanNotifications(rows *sql.Rows) ([]models.Notification, error) {
	var notifications []models.Notification

	for rows.Next() {
		var notification models.Notification
		var createdAtStr, processedAtStr string

		err := rows.Scan(
			&notification.ID,
			&notification.StudentID,
			&notification.Subject,
			&notification.CurrentHours,
			&notification.Threshold,
			&notification.Status,
			&createdAtStr,
			&processedAtStr,
			&notification.StudentName,
		)

		if err != nil {
			return nil, err
		}

		notification.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if processedAtStr != "" {
			notification.ProcessedAt, _ = time.Parse(time.RFC3339, processedAtStr)
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}
