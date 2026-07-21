package api

import (
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

type NotificationAPI struct {
	service *services.NotificationService
}

func NewNotificationAPI() *NotificationAPI {
	return &NotificationAPI{
		service: services.NewNotificationService(),
	}
}

func (api *NotificationAPI) GetNotifications(status string, page, pageSize int) (models.PaginatedResponse, error) {
	return api.service.GetNotifications(status, page, pageSize)
}

func (api *NotificationAPI) GetUnreadCount() (int, error) {
	return api.service.GetUnreadCount()
}

func (api *NotificationAPI) MarkAsRead(id int64) (bool, error) {
	err := api.service.MarkAsRead(id)
	return err == nil, err
}

func (api *NotificationAPI) BatchMarkAsRead(ids []int64) (bool, error) {
	err := api.service.BatchMarkAsRead(ids)
	return err == nil, err
}

func (api *NotificationAPI) BatchDeleteNotifications(ids []int64) (int, error) {
	return api.service.BatchDeleteNotifications(ids)
}

func (api *NotificationAPI) CheckThresholds() (bool, error) {
	err := api.service.CheckThresholds()
	return err == nil, err
}
