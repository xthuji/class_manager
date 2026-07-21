package tests

import (
	"testing"

	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

func TestCourseService_CreateWithDefaultThreshold(t *testing.T) {
	resetTables(t)
	svc := services.NewCourseService()

	c, err := svc.CreateCourse(models.CourseCreateRequest{
		Name: "高一数学",
	})
	if err != nil {
		t.Fatalf("CreateCourse failed: %v", err)
	}
	if c.Threshold != 2 {
		t.Fatalf("expected default threshold 2, got %v", c.Threshold)
	}
}

func TestCourseService_CreateWithExplicitThreshold(t *testing.T) {
	resetTables(t)
	svc := services.NewCourseService()

	c, err := svc.CreateCourse(models.CourseCreateRequest{
		Name:      "高一英语",
		Threshold: 5,
	})
	if err != nil {
		t.Fatalf("CreateCourse failed: %v", err)
	}
	if c.Threshold != 5 {
		t.Fatalf("expected threshold 5, got %v", c.Threshold)
	}
}

func TestCourseService_CreateDuplicateName(t *testing.T) {
	resetTables(t)
	svc := services.NewCourseService()

	if _, err := svc.CreateCourse(models.CourseCreateRequest{Name: "高二数学"}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.CreateCourse(models.CourseCreateRequest{Name: "高二数学"})
	if err == nil {
		t.Fatalf("expected error for duplicate course name, got nil")
	}
}

func TestCourseService_UpdateDuplicateName(t *testing.T) {
	resetTables(t)
	svc := services.NewCourseService()

	c1, _ := svc.CreateCourse(models.CourseCreateRequest{Name: "课程A"})
	_, _ = svc.CreateCourse(models.CourseCreateRequest{Name: "课程B"})

	// 将 c1 改名为 "课程B" 应失败
	_, err := svc.UpdateCourse(models.CourseUpdateRequest{ID: c1.ID, Name: "课程B"})
	if err == nil {
		t.Fatalf("expected error updating to duplicate name, got nil")
	}

	// 改回自身原名应成功
	_, err = svc.UpdateCourse(models.CourseUpdateRequest{ID: c1.ID, Name: "课程A"})
	if err != nil {
		t.Fatalf("expected update to same name to succeed, got %v", err)
	}
}

func TestCourseService_EnrollAndGetStudents(t *testing.T) {
	resetTables(t)
	courseSvc := services.NewCourseService()
	studentSvc := services.NewStudentService()

	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "测试课程"})
	s1, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生1"})
	s2, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生2"})

	if err := courseSvc.EnrollStudent(models.EnrollRequest{StudentID: s1.ID, CourseID: c.ID}); err != nil {
		t.Fatalf("EnrollStudent failed: %v", err)
	}
	if err := courseSvc.EnrollStudent(models.EnrollRequest{StudentID: s2.ID, CourseID: c.ID}); err != nil {
		t.Fatalf("EnrollStudent failed: %v", err)
	}

	// 重复 enroll 应被 OR IGNORE 忽略
	if err := courseSvc.EnrollStudent(models.EnrollRequest{StudentID: s1.ID, CourseID: c.ID}); err != nil {
		t.Fatalf("duplicate enroll should be ignored, got %v", err)
	}

	students, err := courseSvc.GetCourseStudents(c.ID)
	if err != nil {
		t.Fatalf("GetCourseStudents failed: %v", err)
	}
	if len(students) != 2 {
		t.Fatalf("expected 2 enrolled students, got %d", len(students))
	}

	// GetCourseByID 应反映 student_count
	got, _ := courseSvc.GetCourseByID(c.ID)
	if got.StudentCount != 2 {
		t.Fatalf("expected student_count 2, got %d", got.StudentCount)
	}

	// Unenroll
	if err := courseSvc.UnenrollStudent(models.EnrollRequest{StudentID: s1.ID, CourseID: c.ID}); err != nil {
		t.Fatalf("UnenrollStudent failed: %v", err)
	}
	students, _ = courseSvc.GetCourseStudents(c.ID)
	if len(students) != 1 {
		t.Fatalf("expected 1 student after unenroll, got %d", len(students))
	}
}

func TestCourseService_GetCoursesByStudent(t *testing.T) {
	resetTables(t)
	courseSvc := services.NewCourseService()
	studentSvc := services.NewStudentService()

	c1, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程1"})
	c2, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程2"})
	s, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生1"})

	courseSvc.EnrollStudent(models.EnrollRequest{StudentID: s.ID, CourseID: c1.ID})
	courseSvc.EnrollStudent(models.EnrollRequest{StudentID: s.ID, CourseID: c2.ID})

	courses, err := courseSvc.GetCoursesByStudent(s.ID)
	if err != nil {
		t.Fatalf("GetCoursesByStudent failed: %v", err)
	}
	if len(courses) != 2 {
		t.Fatalf("expected 2 courses, got %d", len(courses))
	}
}

func TestCourseService_AddCourseHours(t *testing.T) {
	resetTables(t)
	courseSvc := services.NewCourseService()
	studentSvc := services.NewStudentService()

	c, _ := courseSvc.CreateCourse(models.CourseCreateRequest{Name: "课程X"})
	s1, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生1"})
	s2, _ := studentSvc.CreateStudent(models.StudentCreateRequest{Name: "学生2"})

	courseSvc.EnrollStudent(models.EnrollRequest{StudentID: s1.ID, CourseID: c.ID})
	courseSvc.EnrollStudent(models.EnrollRequest{StudentID: s2.ID, CourseID: c.ID})

	if _, err := courseSvc.AddCourseHours(models.CourseHoursRequest{CourseID: c.ID, Hours: 5}); err != nil {
		t.Fatalf("AddCourseHours failed: %v", err)
	}

	got1, _ := studentSvc.GetStudentByID(s1.ID)
	got2, _ := studentSvc.GetStudentByID(s2.ID)
	if got1.TotalHours != 5 || got2.TotalHours != 5 {
		t.Fatalf("expected both students to have 5 hours, got %v and %v", got1.TotalHours, got2.TotalHours)
	}
}

func TestCourseService_Delete(t *testing.T) {
	resetTables(t)
	svc := services.NewCourseService()

	c, _ := svc.CreateCourse(models.CourseCreateRequest{Name: "待删除"})
	if err := svc.DeleteCourse(c.ID); err != nil {
		t.Fatalf("DeleteCourse failed: %v", err)
	}
	got, _ := svc.GetCourseByID(c.ID)
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}
