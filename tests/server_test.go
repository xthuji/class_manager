package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/server"
)

func doRequest(t *testing.T, srv *server.Server, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	return w
}

func decodeJSONBody(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("decode response failed: %v, body=%s", err, w.Body.String())
	}
}

func TestAPI_StudentCRUD(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{
		"name":        "张三",
		"contact":     "13800000000",
		"total_hours": 10,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}
	var created map[string]interface{}
	decodeJSONBody(t, w, &created)
	id := int64(created["id"].(float64))
	if created["name"] != "张三" {
		t.Fatalf("expected name 张三, got %v", created["name"])
	}

	// 获取
	w = doRequest(t, srv, "GET", "/api/students/"+itoa(id), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 列表
	w = doRequest(t, srv, "GET", "/api/students", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list map[string]interface{}
	decodeJSONBody(t, w, &list)
	if total, ok := list["total"].(float64); !ok || total != 1 {
		t.Fatalf("expected 1 student, got %v", list["total"])
	}

	// 更新
	w = doRequest(t, srv, "PUT", "/api/students/"+itoa(id), map[string]interface{}{
		"name":        "张三改",
		"contact":     "13900000000",
		"total_hours": 20,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// 删除
	w = doRequest(t, srv, "POST", "/api/students/batch-delete", map[string]interface{}{"ids": []int64{id}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPI_GenerateStudentID(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/api/students/generate-id", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	decodeJSONBody(t, w, &resp)
	if resp["student_id"] != "STU0001" {
		t.Fatalf("expected STU0001, got %s", resp["student_id"])
	}
}

func TestAPI_CourseCRUDWithThreshold(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 默认阈值
	w := doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{
		"name": "高一数学",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}
	var c1 map[string]interface{}
	decodeJSONBody(t, w, &c1)
	if c1["threshold"].(float64) != 2 {
		t.Fatalf("expected default threshold 2, got %v", c1["threshold"])
	}
	id1 := int64(c1["id"].(float64))

	// 重复名称应失败
	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{
		"name": "高一数学",
	})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for duplicate, got %d", w.Code)
	}

	// 自定义阈值
	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{
		"name":      "高二英语",
		"threshold": 5,
	})
	var c2 map[string]interface{}
	decodeJSONBody(t, w, &c2)
	if c2["threshold"].(float64) != 5 {
		t.Fatalf("expected threshold 5, got %v", c2["threshold"])
	}

	// 更新
	w = doRequest(t, srv, "PUT", "/api/courses/"+itoa(id1), map[string]interface{}{
		"name":      "高一数学进阶",
		"threshold": 3,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// 列表
	w = doRequest(t, srv, "GET", "/api/courses", nil)
	var list map[string]interface{}
	decodeJSONBody(t, w, &list)
	if total, ok := list["total"].(float64); !ok || total != 2 {
		t.Fatalf("expected 2 courses, got %v", list["total"])
	}
}

func TestAPI_CourseEnroll(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建学生
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	// 创建课程
	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// enroll
	w = doRequest(t, srv, "POST", "/api/courses/enroll", map[string]interface{}{
		"student_id": sid,
		"course_id":  cid,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// 课程学生列表
	w = doRequest(t, srv, "GET", "/api/courses/"+itoa(cid)+"/students", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var students []map[string]interface{}
	decodeJSONBody(t, w, &students)
	if len(students) != 1 {
		t.Fatalf("expected 1 enrolled student, got %d", len(students))
	}

	// 学生课程列表
	w = doRequest(t, srv, "GET", "/api/students/"+itoa(sid)+"/courses", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var courses []map[string]interface{}
	decodeJSONBody(t, w, &courses)
	if len(courses) != 1 {
		t.Fatalf("expected 1 course, got %d", len(courses))
	}
}

func TestAPI_HourRecordBatch(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建 2 个学生 + 1 个课程
	var sids []int64
	for _, name := range []string{"学生1", "学生2"} {
		w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{
			"name": name, "total_hours": 10,
		})
		var s map[string]interface{}
		decodeJSONBody(t, w, &s)
		sids = append(sids, int64(s["id"].(float64)))
	}
	w := doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 批量创建课时记录
	w = doRequest(t, srv, "POST", "/api/hour-records/batch", map[string]interface{}{
		"student_ids": sids,
		"course_id":   cid,
		"hours":       2,
		"record_date": "2026-01-01",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}
	var records []map[string]interface{}
	decodeJSONBody(t, w, &records)
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// 列表
	w = doRequest(t, srv, "GET", "/api/hour-records", nil)
	var list map[string]interface{}
	decodeJSONBody(t, w, &list)
	if total, ok := list["total"].(float64); !ok || total != 2 {
		t.Fatalf("expected 2 records in list, got %v", list["total"])
	}
}

func TestAPI_HourRecharge(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{
		"name": "学生", "total_hours": 5,
	})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	// 创建充值
	w = doRequest(t, srv, "POST", "/api/hour-recharges", map[string]interface{}{
		"student_id":    sid,
		"hours":         10,
		"recharge_date": "2026-01-01",
		"description":   "充值",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	// 按学生查询
	w = doRequest(t, srv, "GET", "/api/hour-recharges/by-student/"+itoa(sid), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list []map[string]interface{}
	decodeJSONBody(t, w, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 recharge, got %d", len(list))
	}
}

func TestAPI_ScheduleCRUD(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建课程
	w := doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 创建排课
	w = doRequest(t, srv, "POST", "/api/schedules", map[string]interface{}{
		"course_id":      cid,
		"day_of_week":    2,
		"period":         "evening",
		"start_time":     "19:00",
		"end_time":       "20:00",
		"hours_consumed": 1,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}
	var sched map[string]interface{}
	decodeJSONBody(t, w, &sched)
	sid := int64(sched["id"].(float64))

	// 按课程查询
	w = doRequest(t, srv, "GET", "/api/schedules/by-course/"+itoa(cid), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list []map[string]interface{}
	decodeJSONBody(t, w, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(list))
	}

	// 删除
	w = doRequest(t, srv, "DELETE", "/api/schedules/"+itoa(sid), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPI_Notifications(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建学生
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	// 直接插入通知（绕过系统通知）
	now := db.GetTimestamp()
	db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, created_at) VALUES (?, ?, ?, ?, 'unread', ?)`,
		sid, "数学", 1, 2, now)

	// 未读数
	w = doRequest(t, srv, "GET", "/api/notifications/unread-count", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var cnt map[string]int
	decodeJSONBody(t, w, &cnt)
	if cnt["count"] != 1 {
		t.Fatalf("expected 1 unread, got %d", cnt["count"])
	}

	// 列表
	w = doRequest(t, srv, "GET", "/api/notifications?status=unread", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list map[string]interface{}
	decodeJSONBody(t, w, &list)
	if total, ok := list["total"].(float64); !ok || total != 1 {
		t.Fatalf("expected 1 notification, got %v", list["total"])
	}

	// 获取通知ID并标记已读
	var notifID int64
	if data, ok := list["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			notifID = int64(item["id"].(float64))
		}
	}
	w = doRequest(t, srv, "POST", "/api/notifications/"+itoa(notifID)+"/read", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPI_DashboardStats(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建 1 学生 + 1 课程 + 充值 + 课时记录（确保课程统计不被过滤）
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 创建充值记录
	doRequest(t, srv, "POST", "/api/hour-recharges", map[string]interface{}{
		"student_id":    sid,
		"hours":         10,
		"recharge_date": time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
	})

	// 创建课时记录（使用近30天内的日期，确保 daily_by_course 查询能命中）
	doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id":  sid,
		"course_id":   cid,
		"hours":       2,
		"record_date": time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
	})

	w = doRequest(t, srv, "GET", "/api/dashboard/stats", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var stats map[string]interface{}
	decodeJSONBody(t, w, &stats)

	// 按课程分组的课时统计
	courseStats, ok := stats["course_stats"].([]interface{})
	if !ok {
		t.Fatalf("expected course_stats array, got %v", stats["course_stats"])
	}
	if len(courseStats) != 1 {
		t.Fatalf("expected 1 course in course_stats, got %d", len(courseStats))
	}

	// 近30天每日按课程分组的消耗课时
	daily, ok := stats["daily_by_course"].([]interface{})
	if !ok {
		t.Fatalf("expected daily_by_course array, got %v", stats["daily_by_course"])
	}
	if len(daily) != 1 {
		t.Fatalf("expected 1 daily entry, got %d", len(daily))
	}
}

func TestAPI_StaticIndexServed(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<html") && !strings.Contains(body, "<!DOCTYPE") {
		t.Fatalf("expected HTML response, got: %s", body[:minInt(200, len(body))])
	}
}

func TestAPI_UnknownAPIPathReturns404(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/api/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAPI_DataExportImport(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 准备数据：学生 + 课程 + 选课关系 + 课时记录
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 选课
	doRequest(t, srv, "POST", "/api/courses/enroll", map[string]interface{}{
		"student_id": sid, "course_id": cid,
	})

	// 课时记录（会生成操作日志）
	doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id": sid, "course_id": cid, "hours": 2, "record_date": "2026-07-20",
	})

	// 验证 student_course 表有数据
	var scCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM student_course`).Scan(&scCount)
	if scCount != 1 {
		t.Fatalf("expected 1 student_course record, got %d", scCount)
	}

	// 验证 operation_logs 表有数据
	var logCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM operation_logs`).Scan(&logCount)
	if logCount == 0 {
		t.Fatal("expected operation_logs records, got 0")
	}

	// 导出
	w = doRequest(t, srv, "GET", "/api/data/export", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var exportResp map[string]string
	decodeJSONBody(t, w, &exportResp)
	data := exportResp["data"]
	if data == "" {
		t.Fatal("expected non-empty export data")
	}

	// 验证导出数据包含 student_course 和 operation_logs
	if !strings.Contains(data, "student_course") {
		t.Fatal("expected export data to contain student_course")
	}
	if !strings.Contains(data, "operation_logs") {
		t.Fatal("expected export data to contain operation_logs")
	}

	// 清空数据库
	resetTables(t)
	count1, _ := getRequest(t, srv, "/api/students")
	if count1 != 0 {
		t.Fatalf("expected 0 students after reset, got %d", count1)
	}

	// 验证 student_course 也被清空
	db.DB.QueryRow(`SELECT COUNT(*) FROM student_course`).Scan(&scCount)
	if scCount != 0 {
		t.Fatalf("expected 0 student_course after reset, got %d", scCount)
	}

	// 导入
	w = doRequest(t, srv, "POST", "/api/data/import", map[string]string{"data": data})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// 验证恢复：学生
	count2, _ := getRequest(t, srv, "/api/students")
	if count2 != 1 {
		t.Fatalf("expected 1 student after import, got %d", count2)
	}

	// 验证恢复：student_course
	db.DB.QueryRow(`SELECT COUNT(*) FROM student_course`).Scan(&scCount)
	if scCount != 1 {
		t.Fatalf("expected 1 student_course after import, got %d", scCount)
	}

	// 验证恢复：operation_logs
	db.DB.QueryRow(`SELECT COUNT(*) FROM operation_logs`).Scan(&logCount)
	if logCount == 0 {
		t.Fatal("expected operation_logs after import, got 0")
	}
}

func getRequest(t *testing.T, srv *server.Server, path string) (int, error) {
	w := doRequest(t, srv, "GET", path, nil)
	if w.Code != http.StatusOK {
		return 0, nil
	}
	var list map[string]interface{}
	decodeJSONBody(t, w, &list)
	if total, ok := list["total"].(float64); ok {
		return int(total), nil
	}
	return 0, nil
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ==================== Dashboard API Tests ====================

func TestAPI_DashboardCharts(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/api/dashboard/charts", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	decodeJSONBody(t, w, &result)
	if _, ok := result["charts"]; !ok {
		t.Fatal("expected charts in response")
	}
}

func TestAPI_DashboardStudentChart(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/api/dashboard/student-chart", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	decodeJSONBody(t, w, &result)
	if _, ok := result["data"]; !ok {
		t.Fatal("expected data in response")
	}
}

// ==================== Student API Tests ====================

func TestAPI_StudentSearch(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "张三"})
	doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "李四"})

	w := doRequest(t, srv, "POST", "/api/students/search", map[string]interface{}{"name": "张"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list map[string]interface{}
	decodeJSONBody(t, w, &list)
	if total, ok := list["total"].(float64); !ok || total != 1 {
		t.Fatalf("expected 1 result, got %v", list["total"])
	}
}

func TestAPI_StudentBatchImport(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "POST", "/api/students/batch-import", map[string]string{
		"csv": "姓名,学号,联系方式,初始课时\n测试1,STU001,13800000001,10\n测试2,STU002,13800000002,20",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	decodeJSONBody(t, w, &result)
	if result["success_count"].(float64) != 2 {
		t.Fatalf("expected 2 success, got %v", result["success_count"])
	}
}

func TestAPI_StudentExport(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "导出测试"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/students/export", map[string]interface{}{"ids": []int64{sid}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]string
	decodeJSONBody(t, w, &result)
	if result["csv"] == "" {
		t.Fatal("expected non-empty csv")
	}
}

// ==================== Course API Tests ====================

func TestAPI_CourseUnenroll(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// Create student and course
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// Enroll
	doRequest(t, srv, "POST", "/api/courses/enroll", map[string]interface{}{
		"student_id": sid,
		"course_id":  cid,
	})

	// Unenroll
	w = doRequest(t, srv, "POST", "/api/courses/unenroll", map[string]interface{}{
		"student_id": sid,
		"course_id":  cid,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify unenrolled
	w = doRequest(t, srv, "GET", "/api/courses/"+itoa(cid)+"/students", nil)
	var students []map[string]interface{}
	decodeJSONBody(t, w, &students)
	if len(students) != 0 {
		t.Fatalf("expected 0 students after unenroll, got %d", len(students))
	}
}

func TestAPI_CourseAddHours(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// Create student and course, enroll
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	doRequest(t, srv, "POST", "/api/courses/enroll", map[string]interface{}{
		"student_id": sid,
		"course_id":  cid,
	})

	// Add hours
	w = doRequest(t, srv, "POST", "/api/courses/"+itoa(cid)+"/add-hours", map[string]interface{}{"hours": 10})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify hours added
	w = doRequest(t, srv, "GET", "/api/students/"+itoa(sid), nil)
	decodeJSONBody(t, w, &s)
	if s["total_hours"].(float64) != 10 {
		t.Fatalf("expected total_hours 10, got %v", s["total_hours"])
	}
}

// ==================== Hour Record API Tests ====================

func TestAPI_HourRecordSingleCreate(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// Create student and course
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生", "total_hours": 10})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// Create single hour record
	w = doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id":  sid,
		"course_id":   cid,
		"hours":       2,
		"record_date": "2026-01-01",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestAPI_HourRecordDates(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/api/hour-records/dates", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var dates []string
	decodeJSONBody(t, w, &dates)
}

func TestAPI_HourRecordSummary(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// Create student, course, recharge, and record
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})

	doRequest(t, srv, "POST", "/api/hour-recharges", map[string]interface{}{
		"student_id":    sid,
		"hours":         10,
		"recharge_date": "2026-01-01",
	})

	w = doRequest(t, srv, "GET", "/api/hour-records/summary/"+itoa(sid), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var summary map[string]interface{}
	decodeJSONBody(t, w, &summary)
	if summary["total_hours"].(float64) != 10 {
		t.Fatalf("expected total_hours 10, got %v", summary["total_hours"])
	}
}

func TestAPI_HourRecordExport(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "POST", "/api/hour-records/export", map[string]interface{}{})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]string
	decodeJSONBody(t, w, &result)
}

func TestAPI_HourRecordUpdate(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// Create student and course
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生", "total_hours": 10})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// Create record
	w = doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id":  sid,
		"course_id":   cid,
		"hours":       2,
		"record_date": "2026-01-01",
	})
	var rec map[string]interface{}
	decodeJSONBody(t, w, &rec)
	rid := int64(rec["id"].(float64))

	// Update record
	w = doRequest(t, srv, "PUT", "/api/hour-records/"+itoa(rid), map[string]interface{}{
		"hours": 3,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPI_HourRecordDelete(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// Create student and course
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生", "total_hours": 10})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// Create record
	w = doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id":  sid,
		"course_id":   cid,
		"hours":       2,
		"record_date": "2026-01-01",
	})
	var rec map[string]interface{}
	decodeJSONBody(t, w, &rec)
	rid := int64(rec["id"].(float64))

	// Delete record
	w = doRequest(t, srv, "POST", "/api/hour-records/batch-delete", map[string]interface{}{"ids": []int64{rid}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ==================== Hour Recharge API Tests ====================

func TestAPI_HourRechargeUpdate(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	// Create recharge
	w = doRequest(t, srv, "POST", "/api/hour-recharges", map[string]interface{}{
		"student_id":    sid,
		"hours":         10,
		"recharge_date": "2026-01-01",
	})
	var rec map[string]interface{}
	decodeJSONBody(t, w, &rec)
	rid := int64(rec["id"].(float64))

	// Update recharge
	w = doRequest(t, srv, "PUT", "/api/hour-recharges/"+itoa(rid), map[string]interface{}{
		"hours":        15,
		"description":  "updated",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ==================== Notification API Tests ====================

func TestAPI_NotificationBatchRead(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建学生
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	// 直接插入一条 unread 通知
	now := db.GetTimestamp()
	res, err := db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, type, created_at) VALUES (?, ?, ?, ?, 'unread', 'threshold', ?)`,
		sid, "数学", 1, 2, now)
	if err != nil {
		t.Fatalf("failed to insert notification: %v", err)
	}
	nid, _ := res.LastInsertId()

	// 批量标记已读
	w = doRequest(t, srv, "POST", "/api/notifications/batch-read", map[string]interface{}{"ids": []int64{nid}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// 验证状态已变为 read
	var status string
	db.DB.QueryRow(`SELECT status FROM notifications WHERE id=?`, nid).Scan(&status)
	if status != "read" {
		t.Fatalf("expected status 'read', got %s", status)
	}
}

func TestAPI_NotificationCheck(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建一个未选课的学生 → 应触发 student_no_course 通知
	doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "未选课学生"})

	// 创建一个没有学生的课程 → 应触发 course_no_students 通知
	doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "无学生课程"})

	// 创建一个选课学生+课程，且课时低于阈值 → 应触发 threshold 通知
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "低课时学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "阈值课程", "threshold": 5})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 设置学生总课时
	db.DB.Exec(`UPDATE students SET total_hours = 10 WHERE id = ?`, sid)

	// 选课
	doRequest(t, srv, "POST", "/api/courses/enroll", map[string]interface{}{
		"student_id": sid, "course_id": cid,
	})

	// 消耗8课时 → 剩余2 < 阈值5
	doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id": sid, "course_id": cid, "hours": 8, "record_date": "2026-07-20",
	})

	// 执行通知检查
	w = doRequest(t, srv, "POST", "/api/notifications/check", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 验证产生了3种类型的通知
	var thresholdCount, studentNoCourseCount, courseNoStudentsCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE type='threshold'`).Scan(&thresholdCount)
	db.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE type='student_no_course'`).Scan(&studentNoCourseCount)
	db.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE type='course_no_students'`).Scan(&courseNoStudentsCount)

	if thresholdCount != 1 {
		t.Fatalf("expected 1 threshold notification, got %d", thresholdCount)
	}
	if studentNoCourseCount != 1 {
		t.Fatalf("expected 1 student_no_course notification, got %d", studentNoCourseCount)
	}
	if courseNoStudentsCount != 1 {
		t.Fatalf("expected 1 course_no_students notification, got %d", courseNoStudentsCount)
	}

	// 再次执行不应产生重复通知
	doRequest(t, srv, "POST", "/api/notifications/check", nil)
	db.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE type='threshold'`).Scan(&thresholdCount)
	if thresholdCount != 1 {
		t.Fatalf("expected 1 threshold notification after re-check, got %d", thresholdCount)
	}
}

func TestAPI_NotificationDelete(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// Create student and insert notification
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	now := db.GetTimestamp()
	res, _ := db.DB.Exec(`INSERT INTO notifications (student_id, course_name, current_hours, threshold, status, type, created_at) VALUES (?, ?, ?, ?, 'unread', 'threshold', ?)`,
		sid, "数学", 1, 2, now)
	nid, _ := res.LastInsertId()

	w = doRequest(t, srv, "POST", "/api/notifications/batch-delete", map[string]interface{}{"ids": []int64{nid}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ==================== Log API Tests ====================

func TestAPI_LogsList(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/api/logs", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	decodeJSONBody(t, w, &result)
	if _, ok := result["total"]; !ok {
		t.Fatal("expected total in response")
	}
}

func TestAPI_LogsEntities(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/api/logs/entities", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []string
	decodeJSONBody(t, w, &result)
	if len(result) == 0 {
		t.Fatal("expected non-empty entities list")
	}
}

func TestAPI_LogsDelete(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	now := db.GetTimestamp()
	res, err := db.DB.Exec(`INSERT INTO operation_logs (entity_type, entity_id, operation_type, description, created_at) VALUES (?, ?, ?, ?, ?)`,
		"student", 1, "create", "test", now)
	if err != nil {
		t.Fatalf("failed to insert log: %v", err)
	}
	lid, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	w := doRequest(t, srv, "POST", "/api/logs/batch-delete", map[string]interface{}{"ids": []int64{lid}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ==================== Data Management API Tests ====================

func TestAPI_DataSaveFile(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "POST", "/api/data/save-file", map[string]interface{}{
		"path":    "test.txt",
		"content": "test content",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// 清理测试生成的文件
	defer os.Remove("test.txt")
}

// ==================== Course Calendar API Tests ====================

func TestAPI_CourseCalendar(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建课程
	w := doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "数学"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 创建学生并选课
	w = doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生A"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	doRequest(t, srv, "POST", "/api/courses/enroll", map[string]interface{}{
		"student_id": sid, "course_id": cid,
	})

	// 创建排课：每周一上午
	doRequest(t, srv, "POST", "/api/schedules", map[string]interface{}{
		"course_id": cid, "day_of_week": 1, "period": "上午",
		"start_time": "09:00", "end_time": "10:00", "hours_consumed": 1,
	})

	// 查询当月日历
	now := time.Now()
	monthStr := now.Format("2006-01")
	w = doRequest(t, srv, "GET", "/api/calendar?month="+monthStr, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	decodeJSONBody(t, w, &result)
	events, ok := result["events"].([]interface{})
	if !ok {
		t.Fatal("expected events array in response")
	}
	if len(events) == 0 {
		t.Fatal("expected at least 1 calendar event")
	}

	// 验证事件包含课程信息
	firstEvent := events[0].(map[string]interface{})
	if firstEvent["course_name"] != "数学" {
		t.Fatalf("expected course_name '数学', got %v", firstEvent["course_name"])
	}
	if firstEvent["period"] != "上午" {
		t.Fatalf("expected period '上午', got %v", firstEvent["period"])
	}

	// 验证事件包含学生列表
	students, ok := firstEvent["students"].([]interface{})
	if !ok || len(students) != 1 {
		t.Fatalf("expected 1 student in event, got %v", firstEvent["students"])
	}
	if students[0] != "学生A" {
		t.Fatalf("expected student '学生A', got %v", students[0])
	}
}

func TestAPI_CourseCalendarEmpty(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 无排课时应返回空事件列表
	w := doRequest(t, srv, "GET", "/api/calendar?month=2026-07", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	decodeJSONBody(t, w, &result)
	events, ok := result["events"].([]interface{})
	if !ok {
		t.Fatal("expected events array in response")
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestAPI_CourseCalendarInvalidMonth(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	w := doRequest(t, srv, "GET", "/api/calendar?month=invalid", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ==================== Dashboard Weekly Consumption Tests ====================

func TestAPI_DashboardWeeklyConsumption(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建学生和课程
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 创建本周的课时记录
	now := time.Now()
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	monday := now.AddDate(0, 0, -daysSinceMonday)
	wednesday := monday.AddDate(0, 0, 2)

	doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id": sid, "course_id": cid, "hours": 3,
		"record_date": monday.Format("2006-01-02"),
	})
	doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id": sid, "course_id": cid, "hours": 2,
		"record_date": wednesday.Format("2006-01-02"),
	})

	// 创建上周的课时记录（不应计入本周）
	lastWeek := monday.AddDate(0, 0, -3)
	doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id": sid, "course_id": cid, "hours": 5,
		"record_date": lastWeek.Format("2006-01-02"),
	})

	// 获取仪表盘统计
	w = doRequest(t, srv, "GET", "/api/dashboard/stats", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var stats map[string]interface{}
	decodeJSONBody(t, w, &stats)

	courseStats, ok := stats["course_stats"].([]interface{})
	if !ok || len(courseStats) != 1 {
		t.Fatalf("expected 1 course in stats, got %v", stats["course_stats"])
	}

	cs := courseStats[0].(map[string]interface{})
	weekHours := cs["this_week_hours"].(float64)
	// 本周应有 3+2=5 课时
	if weekHours != 5 {
		t.Fatalf("expected this_week_hours=5, got %v", weekHours)
	}
}

// ==================== Delete Side-Effect Tests ====================

// TestAPI_HourRecordDeleteReturnsHours 验证删除课时记录会将课时返还给学生（completed_hours 归零）
func TestAPI_HourRecordDeleteReturnsHours(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建学生（total_hours=10）
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{
		"name":        "学生",
		"total_hours": 10,
	})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	// 创建课程
	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 创建课时记录 hours=3
	w = doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id":  sid,
		"course_id":   cid,
		"hours":       3,
		"record_date": "2026-01-01",
	})
	var rec map[string]interface{}
	decodeJSONBody(t, w, &rec)
	rid := int64(rec["id"].(float64))

	// 验证 completed_hours 为 3
	w = doRequest(t, srv, "GET", "/api/students/"+itoa(sid), nil)
	decodeJSONBody(t, w, &s)
	if s["completed_hours"].(float64) != 3 {
		t.Fatalf("expected completed_hours 3, got %v", s["completed_hours"])
	}

	// 删除课时记录
	w = doRequest(t, srv, "POST", "/api/hour-records/batch-delete", map[string]interface{}{"ids": []int64{rid}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// 验证 completed_hours 已归零（课时已返还）
	w = doRequest(t, srv, "GET", "/api/students/"+itoa(sid), nil)
	decodeJSONBody(t, w, &s)
	if s["completed_hours"].(float64) != 0 {
		t.Fatalf("expected completed_hours 0 after delete, got %v", s["completed_hours"])
	}
}

// TestAPI_RechargeDeleteDeductsHours 验证删除充值记录会扣减学生总课时
func TestAPI_RechargeDeleteDeductsHours(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建学生（total_hours 初始为 0）
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	// 创建充值 hours=10
	w = doRequest(t, srv, "POST", "/api/hour-recharges", map[string]interface{}{
		"student_id":    sid,
		"hours":         10,
		"recharge_date": "2026-01-01",
	})
	var rec map[string]interface{}
	decodeJSONBody(t, w, &rec)
	rid := int64(rec["id"].(float64))

	// 验证 total_hours 为 10
	w = doRequest(t, srv, "GET", "/api/students/"+itoa(sid), nil)
	decodeJSONBody(t, w, &s)
	if s["total_hours"].(float64) != 10 {
		t.Fatalf("expected total_hours 10, got %v", s["total_hours"])
	}

	// 删除充值记录
	w = doRequest(t, srv, "POST", "/api/hour-recharges/batch-delete", map[string]interface{}{"ids": []int64{rid}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// 验证 total_hours 已扣减为 0
	w = doRequest(t, srv, "GET", "/api/students/"+itoa(sid), nil)
	decodeJSONBody(t, w, &s)
	if s["total_hours"].(float64) != 0 {
		t.Fatalf("expected total_hours 0 after delete, got %v", s["total_hours"])
	}
}

// TestAPI_BatchDeleteStudents 验证批量删除学生
func TestAPI_BatchDeleteStudents(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建 3 个学生
	var sids []int64
	for i := 1; i <= 3; i++ {
		w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{
			"name": "学生" + itoa(int64(i)),
		})
		var s map[string]interface{}
		decodeJSONBody(t, w, &s)
		sids = append(sids, int64(s["id"].(float64)))
	}

	// 批量删除
	w := doRequest(t, srv, "POST", "/api/students/batch-delete", map[string]interface{}{"ids": sids})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	decodeJSONBody(t, w, &resp)
	if resp["deleted"].(float64) != 3 {
		t.Fatalf("expected deleted 3, got %v", resp["deleted"])
	}

	// 验证 3 个学生全部被删除
	w = doRequest(t, srv, "GET", "/api/students", nil)
	var list map[string]interface{}
	decodeJSONBody(t, w, &list)
	if total, ok := list["total"].(float64); !ok || total != 0 {
		t.Fatalf("expected 0 students after batch delete, got %v", list["total"])
	}
}

// TestAPI_BatchDeleteHourRecords 验证批量删除课时记录并返还课时
func TestAPI_BatchDeleteHourRecords(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建学生和课程
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{
		"name":        "学生",
		"total_hours": 10,
	})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	// 创建 3 条课时记录
	var rids []int64
	for i := 0; i < 3; i++ {
		w := doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
			"student_id":  sid,
			"course_id":   cid,
			"hours":       1,
			"record_date": "2026-01-01",
		})
		var rec map[string]interface{}
		decodeJSONBody(t, w, &rec)
		rids = append(rids, int64(rec["id"].(float64)))
	}

	// 批量删除
	w = doRequest(t, srv, "POST", "/api/hour-records/batch-delete", map[string]interface{}{"ids": rids})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	decodeJSONBody(t, w, &resp)
	if resp["deleted"].(float64) != 3 {
		t.Fatalf("expected deleted 3, got %v", resp["deleted"])
	}

	// 验证 completed_hours 已归零
	w = doRequest(t, srv, "GET", "/api/students/"+itoa(sid), nil)
	decodeJSONBody(t, w, &s)
	if s["completed_hours"].(float64) != 0 {
		t.Fatalf("expected completed_hours 0 after batch delete, got %v", s["completed_hours"])
	}
}

// TestAPI_StudentDeleteCleansUpRelatedData 验证删除学生时清理关联数据
func TestAPI_StudentDeleteCleansUpRelatedData(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建学生
	w := doRequest(t, srv, "POST", "/api/students", map[string]interface{}{"name": "学生"})
	var s map[string]interface{}
	decodeJSONBody(t, w, &s)
	sid := int64(s["id"].(float64))

	// 创建课程并选课
	w = doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{"name": "课程"})
	var c map[string]interface{}
	decodeJSONBody(t, w, &c)
	cid := int64(c["id"].(float64))

	doRequest(t, srv, "POST", "/api/courses/enroll", map[string]interface{}{
		"student_id": sid,
		"course_id":  cid,
	})

	// 创建课时记录
	doRequest(t, srv, "POST", "/api/hour-records", map[string]interface{}{
		"student_id":  sid,
		"course_id":   cid,
		"hours":       2,
		"record_date": "2026-01-01",
	})

	// 创建充值记录
	doRequest(t, srv, "POST", "/api/hour-recharges", map[string]interface{}{
		"student_id":    sid,
		"hours":         10,
		"recharge_date": "2026-01-01",
	})

	// 验证关联数据已存在
	var scCount, hrCount, rhCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM student_course WHERE student_id=?`, sid).Scan(&scCount)
	db.DB.QueryRow(`SELECT COUNT(*) FROM hour_records WHERE student_id=?`, sid).Scan(&hrCount)
	db.DB.QueryRow(`SELECT COUNT(*) FROM hour_recharges WHERE student_id=?`, sid).Scan(&rhCount)
	if scCount != 1 || hrCount != 1 || rhCount != 1 {
		t.Fatalf("expected 1 student_course, 1 hour_record, 1 hour_recharge; got %d, %d, %d", scCount, hrCount, rhCount)
	}

	// 删除学生
	w = doRequest(t, srv, "POST", "/api/students/batch-delete", map[string]interface{}{"ids": []int64{sid}})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// 验证关联数据已被清理
	db.DB.QueryRow(`SELECT COUNT(*) FROM student_course WHERE student_id=?`, sid).Scan(&scCount)
	db.DB.QueryRow(`SELECT COUNT(*) FROM hour_records WHERE student_id=?`, sid).Scan(&hrCount)
	db.DB.QueryRow(`SELECT COUNT(*) FROM hour_recharges WHERE student_id=?`, sid).Scan(&rhCount)
	if scCount != 0 || hrCount != 0 || rhCount != 0 {
		t.Fatalf("expected all related data deleted; got student_course=%d, hour_records=%d, hour_recharges=%d", scCount, hrCount, rhCount)
	}
}

// TestAPI_BatchDeleteCourses 验证批量删除课程
func TestAPI_BatchDeleteCourses(t *testing.T) {
	resetTables(t)
	srv := server.NewServer()

	// 创建 2 个课程
	var cids []int64
	for i := 1; i <= 2; i++ {
		w := doRequest(t, srv, "POST", "/api/courses", map[string]interface{}{
			"name": "课程" + itoa(int64(i)),
		})
		var c map[string]interface{}
		decodeJSONBody(t, w, &c)
		cids = append(cids, int64(c["id"].(float64)))
	}

	// 批量删除
	w := doRequest(t, srv, "POST", "/api/courses/batch-delete", map[string]interface{}{"ids": cids})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	decodeJSONBody(t, w, &resp)
	if resp["deleted"].(float64) != 2 {
		t.Fatalf("expected deleted 2, got %v", resp["deleted"])
	}

	// 验证 2 个课程全部被删除
	w = doRequest(t, srv, "GET", "/api/courses", nil)
	var list map[string]interface{}
	decodeJSONBody(t, w, &list)
	if total, ok := list["total"].(float64); !ok || total != 0 {
		t.Fatalf("expected 0 courses after batch delete, got %v", list["total"])
	}
}
