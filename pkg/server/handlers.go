package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
	"github.com/class_manager/pkg/utils"
)

// --- Dashboard ---

func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	type courseStats struct {
		CourseID          int64   `json:"course_id"`
		CourseName        string  `json:"course_name"`
		TotalHours        float64 `json:"total_hours"`
		ConsumedHours     float64 `json:"consumed_hours"`
		RemainingHours    float64 `json:"remaining_hours"`
		ThisWeekHours     float64 `json:"this_week_hours"`
		ThisMonthHours    float64 `json:"this_month_hours"`
		LastSixMonthsHours float64 `json:"last_six_months_hours"`
		LastYearHours     float64 `json:"last_year_hours"`
	}
	type dailyByCourse struct {
		Date       string  `json:"date"`
		CourseID   int64   `json:"course_id"`
		CourseName string  `json:"course_name"`
		Hours      float64 `json:"hours"`
	}

	// 按课程分组统计（使用子查询避免笛卡尔积）
	courseMap := map[int64]*courseStats{}
	var order []int64
	rows, err := db.DB.Query(`
		SELECT c.id, c.name,
			COALESCE(sc.total_hours, 0) AS total_hours,
			COALESCE(hr.consumed_hours, 0) AS consumed_hours,
			COALESCE(sc.total_hours, 0) - COALESCE(hr.consumed_hours, 0) AS remaining_hours
		FROM courses c
		LEFT JOIN (
			SELECT sc.course_id, SUM(s.total_hours) AS total_hours
			FROM student_course sc
			JOIN students s ON sc.student_id = s.id
			GROUP BY sc.course_id
		) sc ON c.id = sc.course_id
		LEFT JOIN (
			SELECT course_id, SUM(hours) AS consumed_hours
			FROM hour_records
			GROUP BY course_id
		) hr ON c.id = hr.course_id
		ORDER BY c.name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for rows.Next() {
		var cs courseStats
		if err := rows.Scan(&cs.CourseID, &cs.CourseName, &cs.TotalHours, &cs.ConsumedHours, &cs.RemainingHours); err != nil {
			rows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		courseMap[cs.CourseID] = &cs
		order = append(order, cs.CourseID)
	}
	rows.Close()

	// 本周消耗（本周一到今天）
	// SQLite 的 start of week 返回周日，+1 day 得到周一。但周日当天会得到下周一，导致本周数据丢失。
	// 改为在 Go 中计算本周一日期，确保所有情况都正确。
	now := time.Now()
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	monday := now.AddDate(0, 0, -daysSinceMonday).Format("2006-01-02")
	weekRows, err := db.DB.Query(`SELECT course_id, COALESCE(SUM(hours), 0) FROM hour_records 
		WHERE record_date >= ? GROUP BY course_id`, monday)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for weekRows.Next() {
		var courseID int64
		var hours float64
		if err := weekRows.Scan(&courseID, &hours); err != nil {
			weekRows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if cs, ok := courseMap[courseID]; ok {
			cs.ThisWeekHours = hours
		}
	}
	weekRows.Close()

	// 本月消耗（本月初到今天）
	monthRows, err := db.DB.Query(`SELECT course_id, COALESCE(SUM(hours), 0) FROM hour_records 
		WHERE record_date >= date('now', 'start of month') GROUP BY course_id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for monthRows.Next() {
		var courseID int64
		var hours float64
		if err := monthRows.Scan(&courseID, &hours); err != nil {
			monthRows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if cs, ok := courseMap[courseID]; ok {
			cs.ThisMonthHours = hours
		}
	}
	monthRows.Close()

	// 近半年消耗
	halfYearRows, err := db.DB.Query(`SELECT course_id, COALESCE(SUM(hours), 0) FROM hour_records 
		WHERE record_date >= date('now', '-6 months') GROUP BY course_id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for halfYearRows.Next() {
		var courseID int64
		var hours float64
		if err := halfYearRows.Scan(&courseID, &hours); err != nil {
			halfYearRows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if cs, ok := courseMap[courseID]; ok {
			cs.LastSixMonthsHours = hours
		}
	}
	halfYearRows.Close()

	// 近一年消耗
	yearRows, err := db.DB.Query(`SELECT course_id, COALESCE(SUM(hours), 0) FROM hour_records 
		WHERE record_date >= date('now', '-12 months') GROUP BY course_id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for yearRows.Next() {
		var courseID int64
		var hours float64
		if err := yearRows.Scan(&courseID, &hours); err != nil {
			yearRows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if cs, ok := courseMap[courseID]; ok {
			cs.LastYearHours = hours
		}
	}
	yearRows.Close()

	courseStatsList := []courseStats{}
	for _, id := range order {
		cs := *courseMap[id]
		// 过滤掉所有数据都是0的课程
		if cs.ConsumedHours == 0 && cs.ThisWeekHours == 0 && cs.ThisMonthHours == 0 && 
		   cs.LastSixMonthsHours == 0 && cs.LastYearHours == 0 {
			continue
		}
		courseStatsList = append(courseStatsList, cs)
	}

	// 近30天每日按课程分组的消耗课时
	daily := []dailyByCourse{}
	rows2, err := db.DB.Query(`SELECT hr.record_date, c.id, c.name, COALESCE(SUM(hr.hours), 0)
		FROM hour_records hr
		JOIN courses c ON hr.course_id = c.id
		WHERE hr.record_date >= date('now', '-30 days')
		GROUP BY hr.record_date, c.id, c.name
		ORDER BY hr.record_date ASC, c.name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for rows2.Next() {
		var d dailyByCourse
		if err := rows2.Scan(&d.Date, &d.CourseID, &d.CourseName, &d.Hours); err != nil {
			rows2.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		daily = append(daily, d)
	}
	rows2.Close()

	// 汇总统计
	var totalStudents int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&totalStudents); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var totalCourses int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM courses`).Scan(&totalCourses); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var totalHoursConsumed float64
	if err := db.DB.QueryRow(`SELECT COALESCE(SUM(hours), 0) FROM hour_records`).Scan(&totalHoursConsumed); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var totalHours float64
	if err := db.DB.QueryRow(`SELECT COALESCE(SUM(total_hours), 0) FROM students WHERE is_dropped = 0 OR is_dropped IS NULL`).Scan(&totalHours); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_students":       totalStudents,
		"total_courses":        totalCourses,
		"total_hours_consumed": totalHoursConsumed,
		"total_hours":          totalHours,
		"course_stats":         courseStatsList,
		"daily_by_course":      daily,
	})
}

func (s *Server) handleDashboardCharts(w http.ResponseWriter, r *http.Request) {
	type weekEntry struct {
		Week  string  `json:"week"`
		Hours float64 `json:"hours"`
	}
	type monthEntry struct {
		Month string  `json:"month"`
		Hours float64 `json:"hours"`
	}
	type courseChart struct {
		CourseID   int64        `json:"course_id"`
		CourseName string       `json:"course_name"`
		Weeks      []weekEntry  `json:"weeks"`
		Months     []monthEntry `json:"months"`
	}

	// 先获取所有课程，确保每个课程都被包含
	courseMap := map[int64]*courseChart{}
	var order []int64
	rows, err := db.DB.Query(`SELECT id, name FROM courses ORDER BY name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for rows.Next() {
		var courseID int64
		var courseName string
		if err := rows.Scan(&courseID, &courseName); err != nil {
			rows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		courseMap[courseID] = &courseChart{CourseID: courseID, CourseName: courseName}
		order = append(order, courseID)
	}
	rows.Close()

	// 查询近半年按周统计
	rows, err = db.DB.Query(`SELECT c.id, strftime('%Y-W%W', hr.record_date) as week, COALESCE(SUM(hr.hours), 0) FROM courses c LEFT JOIN hour_records hr ON c.id = hr.course_id AND hr.record_date >= date('now', '-6 months') GROUP BY c.id, week ORDER BY c.id, week ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var courseID int64
		var week sql.NullString
		var hours float64
		if err := rows.Scan(&courseID, &week, &hours); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if chart, ok := courseMap[courseID]; ok && week.Valid {
			chart.Weeks = append(chart.Weeks, weekEntry{Week: week.String, Hours: hours})
		}
	}

	// 查询近一年按月统计
	rows2, err := db.DB.Query(`SELECT c.id, strftime('%Y-%m', hr.record_date) as month, COALESCE(SUM(hr.hours), 0) FROM courses c LEFT JOIN hour_records hr ON c.id = hr.course_id AND hr.record_date >= date('now', '-12 months') GROUP BY c.id, month ORDER BY c.id, month ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		var courseID int64
		var month sql.NullString
		var hours float64
		if err := rows2.Scan(&courseID, &month, &hours); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if chart, ok := courseMap[courseID]; ok && month.Valid {
			chart.Months = append(chart.Months, monthEntry{Month: month.String, Hours: hours})
		}
	}

	charts := []courseChart{}
	for _, id := range order {
		chart := *courseMap[id]
		// 过滤掉所有数据都是0的课程（周数据和月数据都为空）
		if len(chart.Weeks) == 0 && len(chart.Months) == 0 {
			continue
		}
		charts = append(charts, chart)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"charts": charts})
}

func (s *Server) handleStudentHourChart(w http.ResponseWriter, r *http.Request) {
	studentIDsStr := r.URL.Query()["student_id"]
	var studentIDs []int64
	for _, s := range studentIDsStr {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid student_id")
			return
		}
		studentIDs = append(studentIDs, id)
	}

	courseIDsStr := r.URL.Query()["course_id"]
	var courseIDs []int64
	for _, s := range courseIDsStr {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid course_id")
			return
		}
		courseIDs = append(courseIDs, id)
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -89).Format("2006-01-02")
	}

	// 获取课程及排课（day_of_week），直接从 courses 和 schedules 获取，不依赖 student_course
	type courseSched struct {
		CourseID   int64
		CourseName string
		DayOfWeek  int
	}
	schedules := []courseSched{}

	// 构建课程 ID 的过滤条件（可选）
	courseFilter := ""
	var courseArgs []interface{}
	if len(courseIDs) > 0 {
		coursePlaceholders := make([]string, len(courseIDs))
		for i, id := range courseIDs {
			coursePlaceholders[i] = "?"
			courseArgs = append(courseArgs, id)
		}
		courseFilter = " WHERE c.id IN (" + strings.Join(coursePlaceholders, ",") + ")"
	}

	query := `SELECT c.id, c.name, sc.day_of_week
		FROM courses c
		JOIN schedules sc ON c.id = sc.course_id` + courseFilter

	rows, err := db.DB.Query(query, courseArgs...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for rows.Next() {
		var cs courseSched
		if err := rows.Scan(&cs.CourseID, &cs.CourseName, &cs.DayOfWeek); err != nil {
			rows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		schedules = append(schedules, cs)
	}
	rows.Close()

	// 获取实际消耗记录（汇总所有学生）
	type consumed struct {
		Date     string
		CourseID int64
		Hours    float64
	}
	consumedMap := map[string]float64{}

	// 构建第二个查询的学生过滤条件
	var consumedStudentFilter string
	var studentArgs []interface{}
	if len(studentIDs) > 0 {
		consumedPlaceholders := make([]string, len(studentIDs))
		for i, id := range studentIDs {
			consumedPlaceholders[i] = "?"
			studentArgs = append(studentArgs, id)
		}
		consumedStudentFilter = "student_id IN (" + strings.Join(consumedPlaceholders, ",") + ")"
	} else {
		consumedStudentFilter = "1=1" // 如果没有指定学生ID，返回所有学生
	}

	consumedQuery := `SELECT record_date, course_id, COALESCE(SUM(hours), 0)
		FROM hour_records
		WHERE ` + consumedStudentFilter + ` AND record_date >= ? AND record_date <= ?`
	if len(courseIDs) > 0 {
		coursePlaceholders := make([]string, len(courseIDs))
		for i := range coursePlaceholders {
			coursePlaceholders[i] = "?"
		}
		consumedQuery += " AND course_id IN (" + strings.Join(coursePlaceholders, ",") + ")"
	}
	consumedQuery += " GROUP BY record_date, course_id"

	consumedArgs := append(studentArgs, startDate, endDate)
	consumedArgs = append(consumedArgs, courseArgs...)

	rows2, err := db.DB.Query(consumedQuery, consumedArgs...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for rows2.Next() {
		var c consumed
		if err := rows2.Scan(&c.Date, &c.CourseID, &c.Hours); err != nil {
			rows2.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		consumedMap[fmt.Sprintf("%s_%d", c.Date, c.CourseID)] = c.Hours
	}
	rows2.Close()

	// 根据排课计算所有应上课日期，合并实际消耗数据
	type chartEntry struct {
		Date       string  `json:"date"`
		CourseID   int64   `json:"course_id"`
		CourseName string  `json:"course_name"`
		Hours      float64 `json:"hours"`
		IsMakeup   bool    `json:"is_makeup"`
	}
	result := []chartEntry{}
	// 记录已存在的(date,course_id)组合，用于过滤补课条目
	existingEntries := map[string]bool{}

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		weekday := int(d.Weekday())
		for _, cs := range schedules {
			if cs.DayOfWeek == weekday {
				key := fmt.Sprintf("%s_%d", d.Format("2006-01-02"), cs.CourseID)
				existingEntries[key] = true
				hours := consumedMap[key]
				result = append(result, chartEntry{
					Date:       d.Format("2006-01-02"),
					CourseID:   cs.CourseID,
					CourseName: cs.CourseName,
					Hours:      hours,
					IsMakeup:   false,
				})
			}
		}
	}

	// 补充非排课时间的补课记录
	type makeupRecord struct {
		Date       string
		CourseID   int64
		CourseName string
		Hours      float64
	}
	makeupQuery := `SELECT hr.record_date, hr.course_id, c.name, COALESCE(SUM(hr.hours), 0)
		FROM hour_records hr
		JOIN courses c ON hr.course_id = c.id
		WHERE ` + consumedStudentFilter + ` AND hr.record_date >= ? AND hr.record_date <= ?`
	if len(courseIDs) > 0 {
		coursePlaceholders := make([]string, len(courseIDs))
		for i := range coursePlaceholders {
			coursePlaceholders[i] = "?"
		}
		makeupQuery += " AND hr.course_id IN (" + strings.Join(coursePlaceholders, ",") + ")"
	}
	makeupQuery += " GROUP BY hr.record_date, hr.course_id, c.name"

	makeupArgs := append(studentArgs, startDate, endDate)
	makeupArgs = append(makeupArgs, courseArgs...)

	makeupRows, err := db.DB.Query(makeupQuery, makeupArgs...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer makeupRows.Close()

	for makeupRows.Next() {
		var mr makeupRecord
		if err := makeupRows.Scan(&mr.Date, &mr.CourseID, &mr.CourseName, &mr.Hours); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		key := fmt.Sprintf("%s_%d", mr.Date, mr.CourseID)
		if !existingEntries[key] {
			result = append(result, chartEntry{
				Date:       mr.Date,
				CourseID:   mr.CourseID,
				CourseName: mr.CourseName,
				Hours:      mr.Hours,
				IsMakeup:   true,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

// --- Course Calendar ---

func (s *Server) handleCourseCalendar(w http.ResponseWriter, r *http.Request) {
	monthStr := r.URL.Query().Get("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	monthTime, err := time.Parse("2006-01", monthStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid month format, use YYYY-MM")
		return
	}

	// 获取该月第一天和下月第一天
	firstOfMonth := time.Date(monthTime.Year(), monthTime.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := firstOfMonth.AddDate(0, 1, 0)

	// 查询所有课程的排课信息
	type scheduleInfo struct {
		CourseID      int64
		CourseName    string
		DayOfWeek     int
		Period        string
		StartTime     string
		EndTime       string
	}
	schedules := []scheduleInfo{}

	rows, err := db.DB.Query(`
		SELECT s.course_id, c.name, s.day_of_week, s.period, s.start_time, s.end_time
		FROM schedules s
		JOIN courses c ON s.course_id = c.id
		ORDER BY c.name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for rows.Next() {
		var si scheduleInfo
		var dayOfWeek sql.NullInt64
		var period, startTime, endTime sql.NullString
		if err := rows.Scan(&si.CourseID, &si.CourseName, &dayOfWeek, &period, &startTime, &endTime); err != nil {
			rows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if dayOfWeek.Valid {
			si.DayOfWeek = int(dayOfWeek.Int64)
		}
		if period.Valid {
			si.Period = period.String
		}
		if startTime.Valid {
			si.StartTime = startTime.String
		}
		if endTime.Valid {
			si.EndTime = endTime.String
		}
		schedules = append(schedules, si)
	}
	rows.Close()

	// 查询每个课程已选课的学生
	courseStudents := map[int64][]string{}
	studentRows, err := db.DB.Query(`
		SELECT sc.course_id, s.name
		FROM student_course sc
		JOIN students s ON sc.student_id = s.id
		ORDER BY s.name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for studentRows.Next() {
		var courseID int64
		var studentName string
		if err := studentRows.Scan(&courseID, &studentName); err != nil {
			studentRows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		courseStudents[courseID] = append(courseStudents[courseID], studentName)
	}
	studentRows.Close()

	// 查询该月有课时记录的学生（按课程+日期分组），用于已上课学生的展示
	type attendanceKey struct {
		CourseID int64
		Date     string
	}
	attendedStudents := map[attendanceKey][]string{}
	attendanceRows, err := db.DB.Query(`
		SELECT hr.course_id, hr.record_date, s.name
		FROM hour_records hr
		JOIN students s ON hr.student_id = s.id
		WHERE hr.record_date >= ? AND hr.record_date < ?
		ORDER BY s.name`, firstOfMonth.Format("2006-01-02"), nextMonth.Format("2006-01-02"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for attendanceRows.Next() {
		var courseID int64
		var recordDate, studentName string
		if err := attendanceRows.Scan(&courseID, &recordDate, &studentName); err != nil {
			attendanceRows.Close()
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		key := attendanceKey{CourseID: courseID, Date: recordDate}
		attendedStudents[key] = append(attendedStudents[key], studentName)
	}
	attendanceRows.Close()

	// 根据排课的 day_of_week 计算该月每个日期的课程
	type calendarEntry struct {
		Date             string   `json:"date"`
		CourseID         int64    `json:"course_id"`
		CourseName       string   `json:"course_name"`
		Period           string   `json:"period"`
		StartTime        string   `json:"start_time"`
		EndTime          string   `json:"end_time"`
		Students         []string `json:"students"`
		AttendedStudents []string `json:"attended_students"`
		IsMakeup         bool     `json:"is_makeup"`
	}

	result := []calendarEntry{}
	// 记录已存在的(date,course_id)组合，用于过滤补课条目
	existingEntries := map[string]bool{}

	for d := firstOfMonth; d.Before(nextMonth); d = d.AddDate(0, 0, 1) {
		weekday := int(d.Weekday()) // 0=Sunday, 1=Monday, ...
		dateStr := d.Format("2006-01-02")
		for _, si := range schedules {
			if si.DayOfWeek == weekday {
				key := fmt.Sprintf("%s_%d", dateStr, si.CourseID)
				existingEntries[key] = true
				result = append(result, calendarEntry{
					Date:             dateStr,
					CourseID:         si.CourseID,
					CourseName:       si.CourseName,
					Period:           si.Period,
					StartTime:        si.StartTime,
					EndTime:          si.EndTime,
					Students:         courseStudents[si.CourseID],
					AttendedStudents: attendedStudents[attendanceKey{CourseID: si.CourseID, Date: dateStr}],
					IsMakeup:         false,
				})
			}
		}
	}

	// 补充非排课时间的补课记录
	type makeupRecord struct {
		Date       string
		CourseID   int64
		CourseName string
	}
	makeupRows, err := db.DB.Query(`
		SELECT DISTINCT hr.record_date, hr.course_id, c.name
		FROM hour_records hr
		JOIN courses c ON hr.course_id = c.id
		WHERE hr.record_date >= ? AND hr.record_date < ?`, firstOfMonth.Format("2006-01-02"), nextMonth.Format("2006-01-02"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer makeupRows.Close()

	for makeupRows.Next() {
		var mr makeupRecord
		if err := makeupRows.Scan(&mr.Date, &mr.CourseID, &mr.CourseName); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		key := fmt.Sprintf("%s_%d", mr.Date, mr.CourseID)
		if !existingEntries[key] {
			result = append(result, calendarEntry{
				Date:             mr.Date,
				CourseID:         mr.CourseID,
				CourseName:       mr.CourseName,
				Period:           "",
				StartTime:        "",
				EndTime:          "",
				Students:         courseStudents[mr.CourseID],
				AttendedStudents: attendedStudents[attendanceKey{CourseID: mr.CourseID, Date: mr.Date}],
				IsMakeup:         true,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"month":  monthStr,
		"events": result,
	})
}

func (s *Server) handleHolidays(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rv := recover(); rv != nil {
			utils.LogErrorf("handleHolidays panic: %v", rv)
			writeError(w, http.StatusInternalServerError, "节假日服务内部错误")
		}
	}()

	monthStr := r.URL.Query().Get("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	svc := services.GetHolidayService()
	holidays, err := svc.GetMonthHolidays(monthStr)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"month":    monthStr,
			"holidays": map[string]interface{}{},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"month":    monthStr,
		"holidays": holidays,
	})
}

// --- Students ---

func (s *Server) handleListStudents(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	result, err := s.studentAPI.ListStudents(models.StudentListRequest{Page: page, PageSize: pageSize})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCreateStudent(w http.ResponseWriter, r *http.Request) {
	var req models.StudentCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.studentAPI.CreateStudent(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleGetStudent(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	result, err := s.studentAPI.GetStudentByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleUpdateStudent(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req models.StudentUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.ID = id
	result, err := s.studentAPI.UpdateStudent(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBatchDeleteStudents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids is required")
		return
	}
	count, err := s.studentAPI.BatchDeleteStudents(req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "deleted": count})
}

func (s *Server) handleGenerateStudentID(w http.ResponseWriter, r *http.Request) {
	result, err := s.studentAPI.GenerateStudentID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"student_id": result})
}

func (s *Server) handleSearchStudents(w http.ResponseWriter, r *http.Request) {
	var req models.StudentSearchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.studentAPI.SearchStudents(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBatchImportStudents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CSV string `json:"csv"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.studentAPI.BatchImportStudents(req.CSV)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleExportStudents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.studentAPI.ExportStudents(req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"csv": result})
}

// --- Courses ---

func (s *Server) handleListCourses(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	result, err := s.courseAPI.ListCourses(models.CourseListRequest{Page: page, PageSize: pageSize})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCreateCourse(w http.ResponseWriter, r *http.Request) {
	var req models.CourseCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.courseAPI.CreateCourse(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleGetCourse(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	result, err := s.courseAPI.GetCourseByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleUpdateCourse(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req models.CourseUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.ID = id
	result, err := s.courseAPI.UpdateCourse(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBatchDeleteCourses(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids is required")
		return
	}
	count, err := s.courseAPI.BatchDeleteCourses(req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "deleted": count})
}

func (s *Server) handleGetCourseStudents(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	result, err := s.courseAPI.GetCourseStudents(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetCoursesByStudent(w http.ResponseWriter, r *http.Request) {
	sid, err := parseID(r, "sid")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid student id")
		return
	}
	result, err := s.courseAPI.GetCoursesByStudent(sid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleEnrollStudent(w http.ResponseWriter, r *http.Request) {
	var req models.EnrollRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := s.courseAPI.EnrollStudent(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) handleUnenrollStudent(w http.ResponseWriter, r *http.Request) {
	var req models.EnrollRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := s.courseAPI.UnenrollStudent(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) handleAddCourseHours(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Hours float64 `json:"hours"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.courseAPI.AddCourseHours(models.CourseHoursRequest{CourseID: id, Hours: req.Hours})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Hour Records ---

func (s *Server) handleListHourRecords(w http.ResponseWriter, r *http.Request) {
	req := models.HourRecordListRequest{
		StartDate:    r.URL.Query().Get("start_date"),
		EndDate:      r.URL.Query().Get("end_date"),
		StudentStatus: r.URL.Query().Get("student_status"),
	}
	
	// 分页参数
	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
		req.Page = page
	}
	if pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil {
		req.PageSize = pageSize
	}
	
	if studentIDs := r.URL.Query()["student_ids"]; len(studentIDs) > 0 {
		for _, idStr := range studentIDs {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				req.StudentIDs = append(req.StudentIDs, id)
			}
		}
	}
	if courseIDs := r.URL.Query()["course_ids"]; len(courseIDs) > 0 {
		for _, idStr := range courseIDs {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				req.CourseIDs = append(req.CourseIDs, id)
			}
		}
	}
	result, err := s.hourRecordAPI.ListHourRecords(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCreateHourRecord(w http.ResponseWriter, r *http.Request) {
	var req models.HourRecordCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.hourRecordAPI.CreateHourRecord(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleBatchCreateHourRecords(w http.ResponseWriter, r *http.Request) {
	var req models.BatchHourRecordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.hourRecordAPI.BatchCreateHourRecord(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleUpdateHourRecord(w http.ResponseWriter, r *http.Request) {
	var req models.HourRecordUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	req.ID = id
	result, err := s.hourRecordAPI.UpdateHourRecord(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBatchDeleteHourRecords(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids is required")
		return
	}
	count, err := s.hourRecordAPI.BatchDeleteHourRecords(req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "deleted": count})
}

func (s *Server) handleGetDistinctDates(w http.ResponseWriter, r *http.Request) {
	result, err := s.hourRecordAPI.GetDistinctDates()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetStudentHoursSummary(w http.ResponseWriter, r *http.Request) {
	sid, err := parseID(r, "sid")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid student id")
		return
	}
	result, err := s.hourRecordAPI.GetStudentHoursSummary(sid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleExportHourRecords(w http.ResponseWriter, r *http.Request) {
	var req models.ExportRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.hourRecordAPI.ExportHourRecords(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"csv": result})
}

// --- Hour Recharges ---

func (s *Server) handleListRecharges(w http.ResponseWriter, r *http.Request) {
	req := models.HourRechargeListRequest{
		Page:      1,
		PageSize:  100,
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	}
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if v, err := strconv.Atoi(pageStr); err == nil {
			req.Page = v
		}
	}
	if psStr := r.URL.Query().Get("page_size"); psStr != "" {
		if v, err := strconv.Atoi(psStr); err == nil {
			req.PageSize = v
		}
	}
	if sidStr := r.URL.Query().Get("student_id"); sidStr != "" {
		if v, err := strconv.ParseInt(sidStr, 10, 64); err == nil {
			req.StudentID = v
		}
	}
	result, err := s.hourRechargeAPI.ListRecharges(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCreateRecharge(w http.ResponseWriter, r *http.Request) {
	var req models.HourRechargeCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RechargeDate == "" {
		req.RechargeDate = time.Now().Format("2006-01-02")
	}
	result, err := s.hourRechargeAPI.CreateRecharge(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleGetRechargesByStudent(w http.ResponseWriter, r *http.Request) {
	sid, err := parseID(r, "sid")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid student id")
		return
	}
	result, err := s.hourRechargeAPI.GetRechargesByStudent(sid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBatchDeleteRecharges(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids is required")
		return
	}
	count, err := s.hourRechargeAPI.BatchDeleteRecharges(req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "deleted": count})
}

func (s *Server) handleUpdateRecharge(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req models.HourRechargeUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.ID = id
	result, err := s.hourRechargeAPI.UpdateRecharge(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Schedules ---

func (s *Server) handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req models.ScheduleCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := s.scheduleAPI.CreateSchedule(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleGetCourseSchedule(w http.ResponseWriter, r *http.Request) {
	cid, err := parseID(r, "cid")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}
	result, err := s.scheduleAPI.GetCourseSchedule(cid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req models.ScheduleUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.ID = id
	result, err := s.scheduleAPI.UpdateSchedule(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	_, err = s.scheduleAPI.DeleteSchedule(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// --- Notifications ---

func (s *Server) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	result, err := s.notificationAPI.GetNotifications(status, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetUnreadCount(w http.ResponseWriter, r *http.Request) {
	count, err := s.notificationAPI.GetUnreadCount()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (s *Server) handleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	_, err = s.notificationAPI.MarkAsRead(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) handleBatchMarkAsRead(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := s.notificationAPI.BatchMarkAsRead(req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) handleCheckThresholds(w http.ResponseWriter, r *http.Request) {
	_, err := s.notificationAPI.CheckThresholds()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) handleBatchDeleteNotifications(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids is required")
		return
	}
	count, err := s.notificationAPI.BatchDeleteNotifications(req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "deleted": count})
}

// --- Data Management ---

func (s *Server) handleExportAllData(w http.ResponseWriter, r *http.Request) {
	result, err := db.ExportAllData()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"data": result})
}

func (s *Server) handleImportAllData(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Data string `json:"data"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err := db.ImportAllData(req.Data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleSaveToFile 将导出的内容写入用户通过原生对话框选择的路径
func (s *Server) handleSaveToFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}
	if err := os.WriteFile(req.Path, []byte(req.Content), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// --- Logs ---

func (s *Server) handleListLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	req := models.OperationLogListRequest{
		Page:           page,
		PageSize:       pageSize,
		EntityType:     r.URL.Query().Get("entity_type"),
		OperationType:  r.URL.Query().Get("operation_type"),
		EntityTypes:    r.URL.Query()["entity_types"],
		OperationTypes: r.URL.Query()["operation_types"],
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
	}

	result, err := s.operationLogAPI.ListLogs(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBatchDeleteLogs(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids is required")
		return
	}
	count, err := s.operationLogAPI.BatchDeleteLogs(req.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "deleted": count})
}

func (s *Server) handleGetLogEntities(w http.ResponseWriter, r *http.Request) {
	entities := []string{"student", "course", "hour_record", "hour_recharge", "schedule", "notification"}
	writeJSON(w, http.StatusOK, entities)
}
