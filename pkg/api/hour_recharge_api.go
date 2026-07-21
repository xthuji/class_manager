package api

import (
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

type HourRechargeAPI struct {
	service *services.HourRechargeService
}

func NewHourRechargeAPI() *HourRechargeAPI {
	return &HourRechargeAPI{service: services.NewHourRechargeService()}
}

func (a *HourRechargeAPI) CreateRecharge(req models.HourRechargeCreateRequest) (*models.HourRecharge, error) {
	return a.service.CreateRecharge(req)
}

func (a *HourRechargeAPI) ListRecharges(req models.HourRechargeListRequest) (models.PaginatedResponse, error) {
	return a.service.ListRecharges(req)
}

func (a *HourRechargeAPI) GetRechargeByID(id int64) (*models.HourRecharge, error) {
	return a.service.GetRechargeByID(id)
}

func (a *HourRechargeAPI) GetRechargesByStudent(studentID int64) ([]models.HourRecharge, error) {
	return a.service.GetRechargesByStudent(studentID)
}

func (a *HourRechargeAPI) BatchDeleteRecharges(ids []int64) (int, error) {
	return a.service.BatchDeleteRecharges(ids)
}

func (a *HourRechargeAPI) UpdateRecharge(req models.HourRechargeUpdateRequest) (*models.HourRecharge, error) {
	return a.service.UpdateRecharge(req)
}
