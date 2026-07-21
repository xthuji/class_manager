package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/class_manager/pkg/models"
	"github.com/class_manager/pkg/services"
)

type HourRechargeAPI struct {
	service *services.HourRechargeService
}

func NewHourRechargeAPI() *HourRechargeAPI {
	return &HourRechargeAPI{service: services.NewHourRechargeService()}
}

func (api *HourRechargeAPI) CreateRecharge(w http.ResponseWriter, r *http.Request) {
	var req models.HourRechargeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recharge, err := api.service.CreateRecharge(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(recharge)
}

func (api *HourRechargeAPI) ListRecharges(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	studentID, _ := strconv.ParseInt(r.URL.Query().Get("student_id"), 10, 64)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	req := models.HourRechargeListRequest{
		Page:     page,
		PageSize: pageSize,
		StudentID: studentID,
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	}

	recharges, err := api.service.ListRecharges(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(recharges)
}

func (api *HourRechargeAPI) GetRecharge(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	recharge, err := api.service.GetRechargeByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if recharge == nil {
		http.Error(w, "Recharge not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(recharge)
}

func (api *HourRechargeAPI) DeleteRecharge(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = api.service.DeleteRecharge(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}