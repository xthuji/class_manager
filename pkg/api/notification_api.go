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

func (api *NotificationAPI) GetNotifications(status string) ([]models.Notification, error) {
	return api.service.GetNotifications(status)
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

func (api *NotificationAPI) MarkAsProcessed(id int64) (bool, error) {
	err := api.service.MarkAsProcessed(id)
	return err == nil, err
}

func (api *NotificationAPI) BatchMarkAsProcessed(ids []int64) (bool, error) {
	err := api.service.BatchMarkAsProcessed(ids)
	return err == nil, err
}

func (api *NotificationAPI) SendSystemNotification(req models.NotificationRequest) (bool, error) {
	err := api.service.SendSystemNotification(req)
	return err == nil, err
}

func (api *NotificationAPI) CheckThresholds() (bool, error) {
	err := api.service.CheckThresholds()
	return err == nil, err
}
