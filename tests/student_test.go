package tests

import (
	"testing"

	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

func TestStudentService_CreateAndList(t *testing.T) {
	resetTables(t)
	svc := services.NewStudentService()

	// 自动生成学号
	s1, err := svc.CreateStudent(models.StudentCreateRequest{
		Name:    "张三",
		Contact: "13800000000",
	})
	if err != nil {
		t.Fatalf("CreateStudent failed: %v", err)
	}
	if s1 == nil || s1.ID == 0 {
		t.Fatalf("expected non-nil student with ID, got %+v", s1)
	}
	if s1.StudentID == "" {
		t.Fatalf("expected auto-generated student_id, got empty")
	}
	if s1.RemainingHours != 0 {
		t.Fatalf("expected remaining hours 0, got %v", s1.RemainingHours)
	}

	// 指定学号
	s2, err := svc.CreateStudent(models.StudentCreateRequest{
		Name:      "李四",
		StudentID: "STU_CUSTOM",
	})
	if err != nil {
		t.Fatalf("CreateStudent with custom id failed: %v", err)
	}
	if s2.StudentID != "STU_CUSTOM" {
		t.Fatalf("expected STU_CUSTOM, got %s", s2.StudentID)
	}

	list, err := svc.ListStudents(models.StudentListRequest{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatalf("ListStudents failed: %v", err)
	}
	if list.Total != 2 {
		t.Fatalf("expected 2 students, got %d", list.Total)
	}
}

func TestStudentService_GenerateStudentID_Increments(t *testing.T) {
	resetTables(t)
	svc := services.NewStudentService()

	id1, err := svc.GenerateStudentID()
	if err != nil {
		t.Fatalf("GenerateStudentID failed: %v", err)
	}
	if id1 != "STU0001" {
		t.Fatalf("expected STU0001, got %s", id1)
	}

	if _, err := svc.CreateStudent(models.StudentCreateRequest{Name: "学生甲"}); err != nil {
		t.Fatalf("CreateStudent failed: %v", err)
	}

	id2, err := svc.GenerateStudentID()
	if err != nil {
		t.Fatalf("GenerateStudentID failed: %v", err)
	}
	if id2 != "STU0002" {
		t.Fatalf("expected STU0002 after 1 student, got %s", id2)
	}
}

func TestStudentService_Update(t *testing.T) {
	resetTables(t)
	svc := services.NewStudentService()

	created, err := svc.CreateStudent(models.StudentCreateRequest{
		Name: "王五",
	})
	if err != nil {
		t.Fatalf("CreateStudent failed: %v", err)
	}

	updated, err := svc.UpdateStudent(models.StudentUpdateRequest{
		ID:         created.ID,
		Name:       "王五改",
		Contact:    "13900000000",
		TotalHours: 15,
	})
	if err != nil {
		t.Fatalf("UpdateStudent failed: %v", err)
	}
	if updated.Name != "王五改" {
		t.Fatalf("expected name 王五改, got %s", updated.Name)
	}
	if updated.TotalHours != 15 {
		t.Fatalf("expected total hours 15, got %v", updated.TotalHours)
	}
}

func TestStudentService_Delete(t *testing.T) {
	resetTables(t)
	svc := services.NewStudentService()

	created, err := svc.CreateStudent(models.StudentCreateRequest{Name: "赵六"})
	if err != nil {
		t.Fatalf("CreateStudent failed: %v", err)
	}

	if err := svc.DeleteStudent(created.ID); err != nil {
		t.Fatalf("DeleteStudent failed: %v", err)
	}

	got, err := svc.GetStudentByID(created.ID)
	if err != nil {
		t.Fatalf("GetStudentByID failed: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}

func TestStudentService_Search(t *testing.T) {
	resetTables(t)
	svc := services.NewStudentService()

	if _, err := svc.CreateStudent(models.StudentCreateRequest{Name: "张三"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateStudent(models.StudentCreateRequest{Name: "李四"}); err != nil {
		t.Fatal(err)
	}

	// 按姓名搜索
	res, err := svc.SearchStudents(models.StudentSearchRequest{Name: "张"})
	if err != nil {
		t.Fatalf("SearchStudents failed: %v", err)
	}
	if res.Total != 1 {
		t.Fatalf("expected 1 result with name 张三, got %+v", res)
	}
}

func TestStudentService_BatchImport(t *testing.T) {
	resetTables(t)
	svc := services.NewStudentService()

	csv := "姓名,学号,联系方式,初始课时\n" +
		"测试1,STU001,13800000001,10\n" +
		"测试2,STU002,13800000002,20\n" +
		",STU003,,5\n" // 姓名为空，应失败

	result, err := svc.BatchImportStudents(csv)
	if err != nil {
		t.Fatalf("BatchImportStudents failed: %v", err)
	}
	if result.SuccessCount != 2 {
		t.Fatalf("expected 2 success, got %d", result.SuccessCount)
	}
	if result.FailedCount != 1 {
		t.Fatalf("expected 1 failed, got %d", result.FailedCount)
	}
}
