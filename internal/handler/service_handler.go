package handler

import (
	"encoding/json"
	"net/http"
	"service-monitor/internal/service"
)

type ServiceHandler struct {
	service *service.Service
}

func NewServiceHandler(service *service.Service) *ServiceHandler {
	return &ServiceHandler{
		service: service,
	}
}

type RegisterServiceRequest struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

type RegisterServiceResponse struct {
	ID int64 `json:"id"`
}

func (h *ServiceHandler) RegisterService(w http.ResponseWriter, r *http.Request) {
	var req RegisterServiceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.service.RegisterService(r.Context(), req.Name, req.Endpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(RegisterServiceResponse{
		ID: id,
	})

}

func (h *ServiceHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.service.GetAllServices(r.Context())
	if err != nil {
		http.Error(w, "failed to get services", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(services); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

}
