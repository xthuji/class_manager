package api

import (
	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

type ThresholdAPI struct {
	service *services.ThresholdService
}

func NewThresholdAPI() *ThresholdAPI {
	return &ThresholdAPI{
		service: services.NewThresholdService(),
	}
}

func (api *ThresholdAPI) SetThreshold(req models.ThresholdRequest) (*models.Threshold, error) {
	return api.service.SetThreshold(req)
}

func (api *ThresholdAPI) GetThreshold(subject string) (*models.Threshold, error) {
	return api.service.GetThreshold(subject)
}

func (api *ThresholdAPI) ListThresholds() ([]models.Threshold, error) {
	return api.service.ListThresholds()
}

func (api *ThresholdAPI) SaveThresholdTemplate(req models.ThresholdTemplateRequest) (bool, error) {
	err := api.service.SaveThresholdTemplate(req)
	return err == nil, err
}

func (api *ThresholdAPI) ApplyThresholdTemplate(templateName string) (bool, error) {
	err := api.service.ApplyThresholdTemplate(templateName)
	return err == nil, err
}

func (api *ThresholdAPI) PreviewThreshold(subject string, value float64) (models.ThresholdPreviewResponse, error) {
	return api.service.PreviewThreshold(subject, value)
}
