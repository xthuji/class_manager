package tests

import (
	"testing"

	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

func TestScheduleService_CRUD(t *testing.T) {
	resetTables(t)
	courseSvc := services.NewCourseService()
	svc := services.NewScheduleService()

	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程"})

	// Create
	sched, err := svc.CreateSchedule(models.ScheduleCreateRequest{
		CourseID:      c.ID,
		DayOfWeek:     2, // Tuesday
		Period:        "evening",
		StartTime:     "19:00",
		EndTime:       "20:00",
		HoursConsumed: 1,
	})
	if err != nil {
		t.Fatalf("CreateSchedule failed: %v", err)
	}
	if sched.ID == 0 {
		t.Fatalf("expected non-zero schedule id")
	}

	// GetCourseSchedule
	list, err := svc.GetCourseSchedule(c.ID)
	if err != nil {
		t.Fatalf("GetCourseSchedule failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(list))
	}

	// Update
	updated, err := svc.UpdateSchedule(models.ScheduleUpdateRequest{
		ID:            sched.ID,
		DayOfWeek:     3,
		Period:        "morning",
		StartTime:     "10:00",
		EndTime:       "11:00",
		HoursConsumed: 1.5,
	})
	if err != nil {
		t.Fatalf("UpdateSchedule failed: %v", err)
	}
	if updated.DayOfWeek != 3 || updated.StartTime != "10:00" {
		t.Fatalf("unexpected updated schedule: %+v", updated)
	}

	// Delete
	if err := svc.DeleteSchedule(sched.ID); err != nil {
		t.Fatalf("DeleteSchedule failed: %v", err)
	}
	list, _ = svc.GetCourseSchedule(c.ID)
	if len(list) != 0 {
		t.Fatalf("expected 0 schedules after delete, got %d", len(list))
	}
}

func TestScheduleService_GetCourseScheduleOrdered(t *testing.T) {
	resetTables(t)
	courseSvc := services.NewCourseService()
	svc := services.NewScheduleService()

	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程"})

	svc.CreateSchedule(models.ScheduleCreateRequest{CourseID: c.ID, DayOfWeek: 6, StartTime: "10:00"})
	svc.CreateSchedule(models.ScheduleCreateRequest{CourseID: c.ID, DayOfWeek: 2, StartTime: "19:00"})

	list, _ := svc.GetCourseSchedule(c.ID)
	if len(list) != 2 {
		t.Fatalf("expected 2 schedules, got %d", len(list))
	}
	// 按 day_of_week 升序排序，星期二(2) 应排在星期六(6) 之前
	if list[0].DayOfWeek != 2 || list[1].DayOfWeek != 6 {
		t.Fatalf("expected ordered by day_of_week, got %d then %d", list[0].DayOfWeek, list[1].DayOfWeek)
	}
}
