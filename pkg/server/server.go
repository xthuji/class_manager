package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/class_manager/pkg/api"
	"github.com/class_manager/pkg/utils"
)

//go:embed static/*
var staticFS embed.FS

// Server HTTP API 服务器
type Server struct {
	studentAPI      *api.StudentAPI
	courseAPI       *api.CourseAPI
	hourRecordAPI   *api.HourRecordAPI
	hourRechargeAPI *api.HourRechargeAPI
	scheduleAPI     *api.ScheduleAPI
	notificationAPI *api.NotificationAPI
	operationLogAPI *api.OperationLogAPI
	mux             *http.ServeMux
}

// NewServer 创建并配置 HTTP 服务器
func NewServer() *Server {
	s := &Server{
		studentAPI:      api.NewStudentAPI(),
		courseAPI:       api.NewCourseAPI(),
		hourRecordAPI:   api.NewHourRecordAPI(),
		hourRechargeAPI: api.NewHourRechargeAPI(),
		scheduleAPI:     api.NewScheduleAPI(),
		notificationAPI: api.NewNotificationAPI(),
		operationLogAPI: api.NewOperationLogAPI(),
		mux:             http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Handler 返回 http.Handler 用于 Wails AssetServer 或独立 HTTP 服务
// 使用 panic recovery 中间件包装，防止任何 handler panic 导致 Wails 应用崩溃
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				utils.LogErrorf("HTTP handler panic: %v\n%s", rv, debug.Stack())
				writeError(w, http.StatusInternalServerError, fmt.Sprintf("内部错误: %v", rv))
			}
		}()
		s.mux.ServeHTTP(w, r)
	})
}

func (s *Server) registerRoutes() {
	// 静态文件
	staticContent, _ := fs.Sub(staticFS, "static")
	fileServer := http.FileServer(http.FS(staticContent))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// API 请求走对应的 handler，其余走静态文件
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// SPA 回退：不存在的路径返回 index.html
		path := r.URL.Path
		if path != "/" && !fileExists(staticContent, path) {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	// Dashboard
	s.mux.HandleFunc("GET /api/dashboard/stats", s.handleDashboardStats)
	s.mux.HandleFunc("GET /api/dashboard/charts", s.handleDashboardCharts)
	s.mux.HandleFunc("GET /api/dashboard/student-chart", s.handleStudentHourChart)

	// Course Calendar
	s.mux.HandleFunc("GET /api/calendar", s.handleCourseCalendar)
	s.mux.HandleFunc("GET /api/holidays", s.handleHolidays)

	// Students
	s.mux.HandleFunc("GET /api/students", s.handleListStudents)
	s.mux.HandleFunc("POST /api/students", s.handleCreateStudent)
	s.mux.HandleFunc("GET /api/students/generate-id", s.handleGenerateStudentID)
	s.mux.HandleFunc("POST /api/students/search", s.handleSearchStudents)
	s.mux.HandleFunc("POST /api/students/batch-import", s.handleBatchImportStudents)
	s.mux.HandleFunc("POST /api/students/export", s.handleExportStudents)
	s.mux.HandleFunc("GET /api/students/{id}", s.handleGetStudent)
	s.mux.HandleFunc("PUT /api/students/{id}", s.handleUpdateStudent)
	s.mux.HandleFunc("POST /api/students/batch-delete", s.handleBatchDeleteStudents)

	// Courses
	s.mux.HandleFunc("GET /api/courses", s.handleListCourses)
	s.mux.HandleFunc("POST /api/courses", s.handleCreateCourse)
	s.mux.HandleFunc("POST /api/courses/enroll", s.handleEnrollStudent)
	s.mux.HandleFunc("POST /api/courses/unenroll", s.handleUnenrollStudent)
	s.mux.HandleFunc("GET /api/courses/{id}", s.handleGetCourse)
	s.mux.HandleFunc("PUT /api/courses/{id}", s.handleUpdateCourse)
	s.mux.HandleFunc("POST /api/courses/batch-delete", s.handleBatchDeleteCourses)
	s.mux.HandleFunc("GET /api/courses/{id}/students", s.handleGetCourseStudents)
	s.mux.HandleFunc("POST /api/courses/{id}/add-hours", s.handleAddCourseHours)

	// Students -> Courses (避免和 /api/courses/{id}/students 冲突)
	s.mux.HandleFunc("GET /api/students/{sid}/courses", s.handleGetCoursesByStudent)

	// Hour Records
	s.mux.HandleFunc("GET /api/hour-records", s.handleListHourRecords)
	s.mux.HandleFunc("POST /api/hour-records", s.handleCreateHourRecord)
	s.mux.HandleFunc("POST /api/hour-records/batch", s.handleBatchCreateHourRecords)
	s.mux.HandleFunc("GET /api/hour-records/dates", s.handleGetDistinctDates)
	s.mux.HandleFunc("GET /api/hour-records/summary/{sid}", s.handleGetStudentHoursSummary)
	s.mux.HandleFunc("POST /api/hour-records/export", s.handleExportHourRecords)
	s.mux.HandleFunc("PUT /api/hour-records/{id}", s.handleUpdateHourRecord)
	s.mux.HandleFunc("POST /api/hour-records/batch-delete", s.handleBatchDeleteHourRecords)

	// Hour Recharges
	s.mux.HandleFunc("GET /api/hour-recharges", s.handleListRecharges)
	s.mux.HandleFunc("POST /api/hour-recharges", s.handleCreateRecharge)
	s.mux.HandleFunc("GET /api/hour-recharges/by-student/{sid}", s.handleGetRechargesByStudent)
	s.mux.HandleFunc("PUT /api/hour-recharges/{id}", s.handleUpdateRecharge)
	s.mux.HandleFunc("POST /api/hour-recharges/batch-delete", s.handleBatchDeleteRecharges)

	// Schedules
	s.mux.HandleFunc("POST /api/schedules", s.handleCreateSchedule)
	s.mux.HandleFunc("GET /api/schedules/by-course/{cid}", s.handleGetCourseSchedule)
	s.mux.HandleFunc("PUT /api/schedules/{id}", s.handleUpdateSchedule)
	s.mux.HandleFunc("DELETE /api/schedules/{id}", s.handleDeleteSchedule)

	// Notifications
	s.mux.HandleFunc("GET /api/notifications", s.handleGetNotifications)
	s.mux.HandleFunc("GET /api/notifications/unread-count", s.handleGetUnreadCount)
	s.mux.HandleFunc("POST /api/notifications/{id}/read", s.handleMarkAsRead)
	s.mux.HandleFunc("POST /api/notifications/batch-read", s.handleBatchMarkAsRead)
	s.mux.HandleFunc("POST /api/notifications/check", s.handleCheckThresholds)
	s.mux.HandleFunc("POST /api/notifications/batch-delete", s.handleBatchDeleteNotifications)

	// Logs
	s.mux.HandleFunc("GET /api/logs", s.handleListLogs)
	s.mux.HandleFunc("GET /api/logs/entities", s.handleGetLogEntities)
	s.mux.HandleFunc("POST /api/logs/batch-delete", s.handleBatchDeleteLogs)

	// Data management
	s.mux.HandleFunc("GET /api/data/export", s.handleExportAllData)
	s.mux.HandleFunc("POST /api/data/import", s.handleImportAllData)
	s.mux.HandleFunc("POST /api/data/save-file", s.handleSaveToFile)
}

// --- 辅助函数 ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseID(r *http.Request, key string) (int64, error) {
	val := r.PathValue(key)
	return strconv.ParseInt(val, 10, 64)
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func fileExists(fsys fs.FS, path string) bool {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return true
	}
	_, err := fs.Stat(fsys, path)
	return err == nil
}

// Start 启动独立 HTTP 服务器
// allowFallback 为 true 时，若配置端口被占用则自动递增寻找可用端口（交付场景）
// allowFallback 为 false 时，严格使用指定端口，端口被占用则直接失败（开发场景）
func (s *Server) Start(port int, allowFallback bool) {
	listener, actualPort, err := listenOrFallback(port, allowFallback)
	if err != nil {
		log.Fatalf("HTTP server failed to start: %v", err)
	}
	log.Printf("HTTP server starting on http://localhost:%d", actualPort)
	if err := http.Serve(listener, s.mux); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

// listenOrFallback 尝试在指定端口监听，若 allowFallback 为 true 则递增端口重试
func listenOrFallback(port int, allowFallback bool) (net.Listener, int, error) {
	maxAttempts := 1
	if allowFallback {
		maxAttempts = 50
	}
	for i := 0; i < maxAttempts; i++ {
		addr := ":" + strconv.Itoa(port + i)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			return listener, port + i, nil
		}
		if !allowFallback {
			return nil, 0, fmt.Errorf("port %d unavailable: %w", port, err)
		}
		log.Printf("Port %d unavailable, trying %d...", port+i, port+i+1)
	}
	return nil, 0, fmt.Errorf("no available port found in range %d-%d", port, port+maxAttempts-1)
}
