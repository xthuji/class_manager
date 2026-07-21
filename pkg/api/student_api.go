package api

import (
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

type StudentAPI struct {
	service *services.StudentService
}

func NewStudentAPI() *StudentAPI {
	return &StudentAPI{
		service: services.NewStudentService(),
	}
}

func (api *StudentAPI) GenerateStudentID() (string, error) {
	return api.service.GenerateStudentID()
}

func (api *StudentAPI) CreateStudent(req models.StudentCreateRequest) (*models.Student, error) {
	return api.service.CreateStudent(req)
}

func (api *StudentAPI) UpdateStudent(req models.StudentUpdateRequest) (*models.Student, error) {
	return api.service.UpdateStudent(req)
}

func (api *StudentAPI) DeleteStudent(id int64) (bool, error) {
	err := api.service.DeleteStudent(id)
	return err == nil, err
}

func (api *StudentAPI) GetStudentByID(id int64) (*models.Student, error) {
	return api.service.GetStudentByID(id)
}

func (api *StudentAPI) ListStudents(req models.StudentListRequest) ([]models.Student, error) {
	return api.service.ListStudents(req)
}

func (api *StudentAPI) SearchStudents(req models.StudentSearchRequest) ([]models.Student, error) {
	return api.service.SearchStudents(req)
}

func (api *StudentAPI) BatchImportStudents(csvContent string) (models.BatchImportResult, error) {
	return api.service.BatchImportStudents(csvContent)
}

func (api *StudentAPI) ExportStudents(ids []int64) (string, error) {
	return api.service.ExportStudents(ids)
}
