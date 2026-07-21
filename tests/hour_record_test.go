package tests

import (
	"testing"

	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

func TestHourRecordService_CreateUpdatesCompletedHours(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	hrSvc := services.NewHourRecordService()
	courseSvc := services.NewCourseService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})
	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程"})
	
	// 先充值10课时
	rechargeSvc := services.NewHourRechargeService()
	rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{
		StudentID:    s.ID,
		Hours:        10,
		RechargeDate: "2026-01-01",
	})

	rec, err := hrSvc.CreateHourRecord(models.HourRecordCreateRequest{
		StudentID:  s.ID,
		CourseID:   c.ID,
		Hours:      2,
		RecordDate: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("CreateHourRecord failed: %v", err)
	}
	if rec.ID == 0 {
		t.Fatalf("expected non-zero record id")
	}

	// completed_hours 应增加 2
	got, _ := studentSvc.GetStudentByID(s.ID)
	if got.CompletedHours != 2 {
		t.Fatalf("expected completed_hours 2, got %v", got.CompletedHours)
	}
	if got.RemainingHours != 8 {
		t.Fatalf("expected remaining 8, got %v", got.RemainingHours)
	}
}

func TestHourRecordService_BatchCreate(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	hrSvc := services.NewHourRecordService()
	courseSvc := services.NewCourseService()

	s1, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生1"})
	s2, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生2"})
	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程"})

	records, err := hrSvc.BatchCreateHourRecord(models.BatchHourRecordRequest{
		StudentIDs: []int64{s1.ID, s2.ID},
		CourseID:   c.ID,
		Hours:      3,
		RecordDate: "2026-01-02",
	})
	if err != nil {
		t.Fatalf("BatchCreateHourRecord failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	got1, _ := studentSvc.GetStudentByID(s1.ID)
	got2, _ := studentSvc.GetStudentByID(s2.ID)
	if got1.CompletedHours != 3 || got2.CompletedHours != 3 {
		t.Fatalf("expected both completed 3, got %v / %v", got1.CompletedHours, got2.CompletedHours)
	}
}

func TestHourRecordService_ListWithFilters(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	hrSvc := services.NewHourRecordService()
	courseSvc := services.NewCourseService()

	s1, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生1"})
	s2, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生2"})
	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程"})

	hrSvc.CreateHourRecord(models.HourRecordCreateRequest{StudentID: s1.ID, CourseID: c.ID, Hours: 1, RecordDate: "2026-01-01"})
	hrSvc.CreateHourRecord(models.HourRecordCreateRequest{StudentID: s2.ID, CourseID: c.ID, Hours: 1, RecordDate: "2026-01-02"})
	hrSvc.CreateHourRecord(models.HourRecordCreateRequest{StudentID: s1.ID, CourseID: c.ID, Hours: 1, RecordDate: "2026-01-03"})

	// 按学生过滤
	res, err := hrSvc.ListHourRecords(models.HourRecordListRequest{StudentIDs: []int64{s1.ID}})
	if err != nil {
		t.Fatalf("ListHourRecords by student failed: %v", err)
	}
	if res.Total != 2 {
		t.Fatalf("expected 2 records for s1, got %d", res.Total)
	}

	// 按日期范围过滤
	res, err = hrSvc.ListHourRecords(models.HourRecordListRequest{StartDate: "2026-01-02", EndDate: "2026-01-02"})
	if err != nil {
		t.Fatalf("ListHourRecords by date failed: %v", err)
	}
	if res.Total != 1 {
		t.Fatalf("expected 1 record in date range, got %d", res.Total)
	}
}

func TestHourRecordService_GetDistinctDates(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	hrSvc := services.NewHourRecordService()
	courseSvc := services.NewCourseService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})
	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程"})

	hrSvc.CreateHourRecord(models.HourRecordCreateRequest{StudentID: s.ID, CourseID: c.ID, Hours: 1, RecordDate: "2026-01-01"})
	hrSvc.CreateHourRecord(models.HourRecordCreateRequest{StudentID: s.ID, CourseID: c.ID, Hours: 1, RecordDate: "2026-01-02"})
	hrSvc.CreateHourRecord(models.HourRecordCreateRequest{StudentID: s.ID, CourseID: c.ID, Hours: 1, RecordDate: "2026-01-01"})

	dates, err := hrSvc.GetDistinctDates()
	if err != nil {
		t.Fatalf("GetDistinctDates failed: %v", err)
	}
	if len(dates) != 2 {
		t.Fatalf("expected 2 distinct dates, got %d", len(dates))
	}
}

func TestHourRecordService_GetStudentHoursSummary(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	hrSvc := services.NewHourRecordService()
	courseSvc := services.NewCourseService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})
	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程"})
	
	// 先充值12课时
	rechargeSvc := services.NewHourRechargeService()
	rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{
		StudentID:    s.ID,
		Hours:        12,
		RechargeDate: "2026-01-01",
	})

	hrSvc.CreateHourRecord(models.HourRecordCreateRequest{StudentID: s.ID, CourseID: c.ID, Hours: 4, RecordDate: "2026-01-01"})

	summary, err := hrSvc.GetStudentHoursSummary(s.ID)
	if err != nil {
		t.Fatalf("GetStudentHoursSummary failed: %v", err)
	}
	if summary.TotalHours != 12 || summary.CompletedHours != 4 || summary.RemainingHours != 8 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
