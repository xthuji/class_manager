package services

import (
	"database/sql"
	"strings"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/utils"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) GetNotifications(status string, page, pageSize int) (models.PaginatedResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query := `SELECT n.id, n.student_id, n.course_name, n.current_hours, n.threshold, n.status, n.type, n.created_at, s.name as student_name
			  FROM notifications n
			  LEFT JOIN students s ON n.student_id = s.id`

	countQuery := `SELECT COUNT(*) FROM notifications n
			  LEFT JOIN students s ON n.student_id = s.id`

	var args []interface{}

	if status != "" {
		query += " WHERE n.status=?"
		countQuery += " WHERE n.status=?"
		args = append(args, status)
	}

	// 获取总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	var total int
	err := db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	query += " ORDER BY n.id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	defer rows.Close()

	notifications, err := s.scanNotifications(rows)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	
	totalPages := (total + pageSize - 1) / pageSize

	return models.PaginatedResponse{
		Data:       notifications,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
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

func (s *NotificationService) CreateNotification(studentID int64, courseName string, currentHours, threshold float64) error {
	return s.createNotification(studentID, courseName, currentHours, threshold, "threshold")
}

func (s *NotificationService) createNotification(studentID int64, courseName string, currentHours, threshold float64, notifType string) error {
	timestamp := db.GetTimestamp()

	query := `INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, type, created_at) VALUES (?, ?, ?, ?, 'unread', ?, ?)`
	_, err := db.DB.Exec(query, studentID, courseName, currentHours, threshold, notifType, timestamp)
	if err != nil {
		return err
	}

	utils.LogInfof("通知创建: 类型=%s, 学生ID=%d, 课程=%s, 剩余=%.1f, 阈值=%.1f", notifType, studentID, courseName, currentHours, threshold)

	return nil
}

func (s *NotificationService) DeleteNotification(id int64) error {
	query := `DELETE FROM notifications WHERE id=?`
	_, err := db.DB.Exec(query, id)
	if err != nil {
		return err
	}
	logService := NewOperationLogService()
	logService.LogChange("delete", "notification", id, "", nil)
	return nil
}

// BatchDeleteNotifications 批量删除通知
func (s *NotificationService) BatchDeleteNotifications(ids []int64) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.DeleteNotification(id); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *NotificationService) CheckThresholds() error {
	// 1. 检查课时阈值
	if err := s.checkHourThresholds(); err != nil {
		return err
	}
	// 2. 检查学生未绑定课程
	if err := s.checkUnboundStudents(); err != nil {
		return err
	}
	// 3. 检查课程没有学生申报
	if err := s.checkCoursesWithoutStudents(); err != nil {
		return err
	}
	return nil
}

func (s *NotificationService) checkHourThresholds() error {
	// 阈值存储在课程级别（courses.threshold），按学生汇总低课时课程
	// 同一学生多个课程低于阈值时，只创建一条通知，不区分课程
	query := `SELECT sc.student_id, s.name, s.total_hours, s.completed_hours, SUM(CASE WHEN (s.total_hours - s.completed_hours) < c.threshold THEN 1 ELSE 0 END) AS low_count
		FROM courses c
		JOIN student_course sc ON c.id = sc.course_id
		JOIN students s ON sc.student_id = s.id
		WHERE c.threshold > 0 AND (s.is_dropped = 0 OR s.is_dropped IS NULL)
		GROUP BY sc.student_id, s.name, s.total_hours, s.completed_hours
		HAVING SUM(CASE WHEN (s.total_hours - s.completed_hours) < c.threshold THEN 1 ELSE 0 END) > 0`

	rows, err := db.DB.Query(query)
	if err != nil {
		return err
	}

	type pendingNotif struct {
		StudentID      int64
		StudentName    string
		TotalHours     float64
		CompletedHours float64
	}
	var pending []pendingNotif

	for rows.Next() {
		var studentID int64
		var studentName string
		var totalHours, completedHours, lowCount float64

		if err := rows.Scan(&studentID, &studentName, &totalHours, &completedHours, &lowCount); err != nil {
			rows.Close()
			return err
		}

		// 同一学生当天已创建过通知则跳过，避免重复提醒
		checkQuery := `SELECT COUNT(*) FROM notifications WHERE student_id=? AND type='threshold' AND date(created_at)=date('now')`
		var exists int
		db.DB.QueryRow(checkQuery, studentID).Scan(&exists)

		if exists == 0 {
			pending = append(pending, pendingNotif{
				StudentID:      studentID,
				StudentName:    studentName,
				TotalHours:     totalHours,
				CompletedHours: completedHours,
			})
		}
	}
	rows.Close()

	for _, p := range pending {
		remainingHours := p.TotalHours - p.CompletedHours
		if err := s.createNotification(p.StudentID, "", remainingHours, 0, "threshold"); err != nil {
			return err
		}
	}

	return nil
}

func (s *NotificationService) checkUnboundStudents() error {
	// 查找未退学但没有绑定任何课程的学生
	query := `SELECT s.id, s.name FROM students s
		WHERE (s.is_dropped = 0 OR s.is_dropped IS NULL)
		AND s.id NOT IN (SELECT DISTINCT student_id FROM student_course)`

	rows, err := db.DB.Query(query)
	if err != nil {
		return err
	}

	type pendingNotif struct {
		StudentID  int64
		StudentName string
	}
	var pending []pendingNotif

	for rows.Next() {
		var studentID int64
		var studentName string
		if err := rows.Scan(&studentID, &studentName); err != nil {
			rows.Close()
			return err
		}

		// 同一学生当天已创建过此类通知则跳过
		checkQuery := `SELECT COUNT(*) FROM notifications WHERE student_id=? AND type='student_no_course' AND date(created_at)=date('now')`
		var exists int
		db.DB.QueryRow(checkQuery, studentID).Scan(&exists)

		if exists == 0 {
			pending = append(pending, pendingNotif{StudentID: studentID, StudentName: studentName})
		}
	}
	rows.Close()

	for _, p := range pending {
		if err := s.createNotification(p.StudentID, "", 0, 0, "student_no_course"); err != nil {
			return err
		}
	}

	return nil
}

func (s *NotificationService) checkCoursesWithoutStudents() error {
	// 查找没有在读（未退学）学生申报的课程：仅统计未退学学生的选课记录
	query := `SELECT c.id, c.name FROM courses c
		WHERE c.id NOT IN (
			SELECT DISTINCT sc.course_id FROM student_course sc
			JOIN students s ON sc.student_id = s.id
			WHERE s.is_dropped = 0 OR s.is_dropped IS NULL
		)`

	rows, err := db.DB.Query(query)
	if err != nil {
		return err
	}

	type pendingNotif struct {
		CourseID   int64
		CourseName string
	}
	var pending []pendingNotif

	for rows.Next() {
		var courseID int64
		var courseName string
		if err := rows.Scan(&courseID, &courseName); err != nil {
			rows.Close()
			return err
		}

		// 同一课程当天已创建过此类通知则跳过
		checkQuery := `SELECT COUNT(*) FROM notifications WHERE course_name=? AND type='course_no_students' AND date(created_at)=date('now')`
		var exists int
		db.DB.QueryRow(checkQuery, courseName).Scan(&exists)

		if exists == 0 {
			pending = append(pending, pendingNotif{CourseID: courseID, CourseName: courseName})
		}
	}
	rows.Close()

	for _, p := range pending {
		// student_id 设为 0 表示无关联学生
		if err := s.createNotification(0, p.CourseName, 0, 0, "course_no_students"); err != nil {
			return err
		}
	}

	return nil
}

func (s *NotificationService) scanNotifications(rows *sql.Rows) ([]models.Notification, error) {
	notifications := []models.Notification{}

	for rows.Next() {
		var notification models.Notification
		var createdAtStr string
		var studentName sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.StudentID,
			&notification.CourseName,
			&notification.CurrentHours,
			&notification.Threshold,
			&notification.Status,
			&notification.Type,
			&createdAtStr,
			&studentName,
		)

		if err != nil {
			return nil, err
		}

		notification.StudentName = studentName.String
		notification.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		notifications = append(notifications, notification)
	}

	return notifications, nil
}
