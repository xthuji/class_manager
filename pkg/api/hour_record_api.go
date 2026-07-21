package api

import (
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

type HourRecordAPI struct {
	service *services.HourRecordService
}

func NewHourRecordAPI() *HourRecordAPI {
	return &HourRecordAPI{
		service: services.NewHourRecordService(),
	}
}

func (api *HourRecordAPI) CreateHourRecord(req models.HourRecordCreateRequest) (*models.HourRecord, error) {
	return api.service.CreateHourRecord(req)
}

func (api *HourRecordAPI) BatchCreateHourRecord(req models.BatchHourRecordRequest) ([]models.HourRecord, error) {
	return api.service.BatchCreateHourRecord(req)
}

func (api *HourRecordAPI) GetHourRecordsByStudent(studentID int64) ([]models.HourRecord, error) {
	return api.service.GetHourRecordsByStudent(studentID)
}

func (api *HourRecordAPI) ExportHourRecords(req models.ExportRequest) (string, error) {
	return api.service.ExportHourRecords(req)
}

func (api *HourRecordAPI) GetStudentHoursSummary(studentID int64) (models.HoursSummaryResponse, error) {
	return api.service.GetStudentHoursSummary(studentID)
}

func (api *HourRecordAPI) ListHourRecords(req models.HourRecordListRequest) ([]models.HourRecordWithDetails, error) {
	return api.service.ListHourRecords(req)
}

func (api *HourRecordAPI) GetDistinctDates() ([]string, error) {
	return api.service.GetDistinctDates()
}
