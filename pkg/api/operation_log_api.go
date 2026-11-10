package api

import (
	"class_manager/pkg/models"
	"class_manager/pkg/services"
)

type OperationLogAPI struct {
	service *services.OperationLogService
}

func NewOperationLogAPI() *OperationLogAPI {
	return &OperationLogAPI{
		service: services.NewOperationLogService(),
	}
}

func (api *OperationLogAPI) ListLogs(req models.OperationLogListRequest) (models.PaginatedResponse, error) {
	return api.service.ListLogs(req)
}

func (api *OperationLogAPI) BatchDeleteLogs(ids []int64) (int, error) {
	return api.service.BatchDeleteLogs(ids)
}
