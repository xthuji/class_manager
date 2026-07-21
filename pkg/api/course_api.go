package api

import (
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

type CourseAPI struct {
	service *services.CourseService
}

func NewCourseAPI() *CourseAPI {
	return &CourseAPI{
		service: services.NewCourseService(),
	}
}

func (api *CourseAPI) CreateCourse(req models.CourseCreateRequest) (*models.Course, error) {
	return api.service.CreateCourse(req)
}

func (api *CourseAPI) UpdateCourse(req models.CourseUpdateRequest) (*models.Course, error) {
	return api.service.UpdateCourse(req)
}

func (api *CourseAPI) BatchDeleteCourses(ids []int64) (int, error) {
	return api.service.BatchDeleteCourses(ids)
}

func (api *CourseAPI) GetCourseByID(id int64) (*models.Course, error) {
	return api.service.GetCourseByID(id)
}

func (api *CourseAPI) ListCourses(req models.CourseListRequest) (models.PaginatedResponse, error) {
	return api.service.ListCourses(req)
}

func (api *CourseAPI) EnrollStudent(req models.EnrollRequest) (bool, error) {
	err := api.service.EnrollStudent(req)
	return err == nil, err
}

func (api *CourseAPI) UnenrollStudent(req models.EnrollRequest) (bool, error) {
	err := api.service.UnenrollStudent(req)
	return err == nil, err
}

func (api *CourseAPI) GetCourseStudents(courseID int64) ([]models.Student, error) {
	return api.service.GetCourseStudents(courseID)
}

func (api *CourseAPI) GetCoursesByStudent(studentID int64) ([]models.Course, error) {
	return api.service.GetCoursesByStudent(studentID)
}

func (api *CourseAPI) AddCourseHours(req models.CourseHoursRequest) (*models.Course, error) {
	return api.service.AddCourseHours(req)
}
