package tests

import (
	"testing"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

func TestNotificationService_GetUnreadCount(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	svc := services.NewNotificationService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})

	// 直接插入通知，绕过 CreateNotification 中的系统通知调用
	now := db.GetTimestamp()
	db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, created_at) VALUES (?, ?, ?, ?, 'unread', ?)`,
		s.ID, "数学", 1, 2, now)
	db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, created_at) VALUES (?, ?, ?, ?, 'read', ?)`,
		s.ID, "数学", 1, 2, now)

	count, err := svc.GetUnreadCount()
	if err != nil {
		t.Fatalf("GetUnreadCount failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 unread, got %d", count)
	}
}

func TestNotificationService_GetByStatus(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	svc := services.NewNotificationService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})
	now := db.GetTimestamp()
	db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, created_at) VALUES (?, ?, ?, ?, 'unread', ?)`,
		s.ID, "数学", 1, 2, now)
	db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, created_at) VALUES (?, ?, ?, ?, 'read', ?)`,
		s.ID, "数学", 1, 2, now)

	// 全部
	all, err := svc.GetNotifications("", 1, 100)
	if err != nil {
		t.Fatalf("GetNotifications all failed: %v", err)
	}
	if all.Total != 2 {
		t.Fatalf("expected 2 total, got %d", all.Total)
	}

	// 只看 unread
	unread, err := svc.GetNotifications("unread", 1, 100)
	if err != nil {
		t.Fatalf("GetNotifications unread failed: %v", err)
	}
	if unread.Total != 1 {
		t.Fatalf("expected 1 unread, got %d", unread.Total)
	}
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	svc := services.NewNotificationService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})
	now := db.GetTimestamp()
	res, _ := db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, created_at) VALUES (?, ?, ?, ?, 'unread', ?)`,
		s.ID, "数学", 1, 2, now)
	id, _ := res.LastInsertId()

	// 标记已读
	if err := svc.MarkAsRead(id); err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}
	all, _ := svc.GetNotifications("unread", 1, 100)
	if all.Total != 0 {
		t.Fatalf("expected 0 unread after mark read, got %d", all.Total)
	}
	read, _ := svc.GetNotifications("read", 1, 100)
	if read.Total != 1 {
		t.Fatalf("expected 1 read, got %d", read.Total)
	}
}

func TestNotificationService_BatchMark(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	svc := services.NewNotificationService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})
	now := db.GetTimestamp()

	var ids []int64
	for i := 0; i < 3; i++ {
		res, _ := db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, created_at) VALUES (?, ?, ?, ?, 'unread', ?)`,
			s.ID, "数学", 1, 2, now)
		id, _ := res.LastInsertId()
		ids = append(ids, id)
	}

	if err := svc.BatchMarkAsRead(ids); err != nil {
		t.Fatalf("BatchMarkAsRead failed: %v", err)
	}
	unread, _ := svc.GetNotifications("unread", 1, 100)
	if unread.Total != 0 {
		t.Fatalf("expected 0 unread after batch, got %d", unread.Total)
	}
	read, _ := svc.GetNotifications("read", 1, 100)
	if read.Total != 3 {
		t.Fatalf("expected 3 read after batch, got %d", read.Total)
	}
}

func TestNotificationService_CheckThresholds(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	svc := services.NewNotificationService()
	courseSvc := services.NewCourseService()

	// 创建课程，阈值设为 5
	c, err := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "数学课", Threshold: 5})
	if err != nil {
		t.Fatalf("create course failed: %v", err)
	}

	// 创建学生，总课时 10
	if _, err := studentSvc.CreateStudent(models.StudentCreateRequest{
		Name: "低课时学生",
	}); err != nil {
		t.Fatalf("create student failed: %v", err)
	}

	students, _ := studentSvc.ListStudents(models.StudentListRequest{Page: 1, PageSize: 100})
	if students.Total == 0 {
		t.Fatal("expected at least 1 student")
	}

	// 直接从数据库获取学生ID
	var studentID int64
	if err := db.DB.QueryRow(`SELECT id FROM students LIMIT 1`).Scan(&studentID); err != nil {
		t.Fatalf("get student id failed: %v", err)
	}

	// 学生选课
	if err := courseSvc.EnrollStudent(models.EnrollRequest{StudentID: studentID, CourseID: c.ID}); err != nil {
		t.Fatalf("enroll failed: %v", err)
	}

	// 直接设置学生总课时为10
	if _, err := db.DB.Exec(`UPDATE students SET total_hours = 10 WHERE id = ?`, studentID); err != nil {
		t.Fatalf("update student hours failed: %v", err)
	}

	// 通过 hour_record 消耗 8 课时，剩 2 < 5（使用最近日期）
	hrSvc := services.NewHourRecordService()
	if _, err := hrSvc.CreateHourRecord(models.HourRecordCreateRequest{
		StudentID:  studentID,
		CourseID:   c.ID,
		Hours:      8,
		RecordDate: "2026-07-20",
	}); err != nil {
		t.Fatalf("create hour record failed: %v", err)
	}

	// 验证学生数据状态
	var total, completed float64
	if err := db.DB.QueryRow(`SELECT total_hours, completed_hours FROM students WHERE id = ?`, studentID).Scan(&total, &completed); err != nil {
		t.Fatalf("query student hours failed: %v", err)
	}
	t.Logf("student total_hours=%.1f, completed_hours=%.1f, remaining=%.1f", total, completed, total-completed)

	if err := svc.CheckThresholds(); err != nil {
		t.Fatalf("CheckThresholds failed: %v", err)
	}

	// 应该有一条未读通知，course_name 为课程名称
	unread, _ := svc.GetNotifications("unread", 1, 100)
	if unread.Total != 1 {
		t.Fatalf("expected 1 unread notification after threshold check, got %d", unread.Total)
	}

	// 再次执行不应产生重复通知
	if err := svc.CheckThresholds(); err != nil {
		t.Fatalf("second CheckThresholds failed: %v", err)
	}
	unread, _ = svc.GetNotifications("unread", 1, 100)
	if unread.Total != 1 {
		t.Fatalf("expected still 1 unread notification, got %d", unread.Total)
	}
}

func TestNotificationService_ExcludeDroppedStudents(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	courseSvc := services.NewCourseService()
	svc := services.NewNotificationService()

	// 创建课程，阈值设为 5
	c, err := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "数学课", Threshold: 5})
	if err != nil {
		t.Fatalf("create course failed: %v", err)
	}

	// 在读低课时学生 → 应触发 threshold 通知
	active, err := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "在读低课时学生"})
	if err != nil {
		t.Fatalf("create active student failed: %v", err)
	}
	// 已退学的低课时选课学生 → 不应触发任何通知
	dropped, err := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "退学选课学生"})
	if err != nil {
		t.Fatalf("create dropped student failed: %v", err)
	}
	// 已退学且未选课学生 → 不应触发 student_no_course 通知
	droppedUnbound, err := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "退学未选课学生"})
	if err != nil {
		t.Fatalf("create dropped unbound student failed: %v", err)
	}

	if _, err := db.DB.Exec(`UPDATE students SET is_dropped = 1 WHERE id IN (?, ?)`, dropped.ID, droppedUnbound.ID); err != nil {
		t.Fatalf("mark students dropped failed: %v", err)
	}

	// 仅在读学生和退学学生选课
	if err := courseSvc.EnrollStudent(models.EnrollRequest{StudentID: active.ID, CourseID: c.ID}); err != nil {
		t.Fatalf("enroll active failed: %v", err)
	}
	if err := courseSvc.EnrollStudent(models.EnrollRequest{StudentID: dropped.ID, CourseID: c.ID}); err != nil {
		t.Fatalf("enroll dropped failed: %v", err)
	}

	if err := svc.CheckThresholds(); err != nil {
		t.Fatalf("CheckThresholds failed: %v", err)
	}

	// 退学学生不应有任何通知
	var cnt int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE student_id = ?`, dropped.ID).Scan(&cnt); err != nil {
		t.Fatalf("query dropped notifications failed: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("expected 0 notification for dropped enrolled student, got %d", cnt)
	}
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE student_id = ?`, droppedUnbound.ID).Scan(&cnt); err != nil {
		t.Fatalf("query dropped unbound notifications failed: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("expected 0 notification for dropped unbound student, got %d", cnt)
	}

	// 在读学生应有一条 threshold 通知
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE student_id = ? AND type = 'threshold'`, active.ID).Scan(&cnt); err != nil {
		t.Fatalf("query active notifications failed: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 threshold notification for active student, got %d", cnt)
	}
}
