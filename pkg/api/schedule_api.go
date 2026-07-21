package api

import (
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

type ScheduleAPI struct {
	service *services.ScheduleService
}

func NewScheduleAPI() *ScheduleAPI {
	return &ScheduleAPI{
		service: services.NewScheduleService(),
	}
}

func (api *ScheduleAPI) CreateSchedule(req models.ScheduleCreateRequest) (*models.Schedule, error) {
	return api.service.CreateSchedule(req)
}

func (api *ScheduleAPI) UpdateSchedule(req models.ScheduleUpdateRequest) (*models.Schedule, error) {
	return api.service.UpdateSchedule(req)
}

func (api *ScheduleAPI) DeleteSchedule(id int64) (bool, error) {
	err := api.service.DeleteSchedule(id)
	return err == nil, err
}

func (api *ScheduleAPI) GetCourseSchedule(courseID int64) ([]models.Schedule, error) {
	return api.service.GetCourseSchedule(courseID)
}
