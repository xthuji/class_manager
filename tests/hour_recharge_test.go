package tests

import (
	"testing"

	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

func TestHourRechargeService_CreateIncreasesTotalHours(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	rechargeSvc := services.NewHourRechargeService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})

	rec, err := rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{
		StudentID:    s.ID,
		Hours:        10,
		RechargeDate: "2026-01-01",
		Description:  "首次充值",
	})
	if err != nil {
		t.Fatalf("CreateRecharge failed: %v", err)
	}
	if rec == nil || rec.ID == 0 {
		t.Fatalf("expected non-nil recharge with ID, got %+v", rec)
	}
	if rec.StudentName != "学生" {
		t.Fatalf("expected student name 学生, got %s", rec.StudentName)
	}

	got, _ := studentSvc.GetStudentByID(s.ID)
	if got.TotalHours != 10 {
		t.Fatalf("expected total hours 10 after recharge, got %v", got.TotalHours)
	}
}

func TestHourRechargeService_ListByStudent(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	rechargeSvc := services.NewHourRechargeService()

	s1, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生1"})
	s2, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生2"})

	rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{StudentID: s1.ID, Hours: 5, RechargeDate: "2026-01-01"})
	rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{StudentID: s1.ID, Hours: 3, RechargeDate: "2026-01-02"})
	rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{StudentID: s2.ID, Hours: 2, RechargeDate: "2026-01-03"})

	list, err := rechargeSvc.GetRechargesByStudent(s1.ID)
	if err != nil {
		t.Fatalf("GetRechargesByStudent failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 recharges for s1, got %d", len(list))
	}
}

func TestHourRechargeService_Delete(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	rechargeSvc := services.NewHourRechargeService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})
	rec, _ := rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{StudentID: s.ID, Hours: 5, RechargeDate: "2026-01-01"})

	if err := rechargeSvc.DeleteRecharge(rec.ID); err != nil {
		t.Fatalf("DeleteRecharge failed: %v", err)
	}

	got, _ := rechargeSvc.GetRechargeByID(rec.ID)
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}

func TestHourRechargeService_ListWithFilters(t *testing.T) {
	resetTables(t)
	studentSvc := services.NewStudentService()
	rechargeSvc := services.NewHourRechargeService()

	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生"})
	rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{StudentID: s.ID, Hours: 5, RechargeDate: "2026-01-01"})
	rechargeSvc.CreateRecharge(models.HourRechargeCreateRequest{StudentID: s.ID, Hours: 3, RechargeDate: "2026-02-01"})

	// 日期范围过滤
	list, err := rechargeSvc.ListRecharges(models.HourRechargeListRequest{
		StudentID: s.ID,
		StartDate: "2026-01-01",
		EndDate:   "2026-01-31",
		Page:      1,
		PageSize:  100,
	})
	if err != nil {
		t.Fatalf("ListRecharges failed: %v", err)
	}
	if list.Total != 1 {
		t.Fatalf("expected 1 recharge in January, got %d", list.Total)
	}
}
