package http

import (
	"net/http"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

// technicianResponse is one row of the technician table, with account active state and availability.
type technicianResponse struct {
	ID                   string `json:"id"`
	UserID               string `json:"userId"`
	FullName             string `json:"fullName"`
	Specialty            string `json:"specialty"`
	IsActive             bool   `json:"isActive"`
	Busy                 bool   `json:"busy"`
	CanReceiveAssignment bool   `json:"canReceiveAssignment"`
	ActiveOrderID        string `json:"activeOrderId"`
	ActiveOrderNumber    string `json:"activeOrderNumber"`
	ActiveVehiclePlate   string `json:"activeVehiclePlate"`
}

type createTechnicianRequest struct {
	FullName  string `json:"fullName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Specialty string `json:"specialty"`
}

type setAccessRequest struct {
	Active bool `json:"active"`
}

// TechnicianHandler exposes technician allocation, onboarding, and access control.
type TechnicianHandler struct {
	technician usecase.TechnicianUseCase
}

// NewTechnicianHandler wires the technician handler.
func NewTechnicianHandler(technician usecase.TechnicianUseCase) TechnicianHandler {
	return TechnicianHandler{technician: technician}
}

// List returns every technician with the order they currently hold and account active status.
func (h TechnicianHandler) List(writer http.ResponseWriter, request *http.Request) {
	if _, err := requireAdministrator(request.Context()); err != nil {
		failure(writer, err)
		return
	}
	listed, err := h.technician.ListWorkload(request.Context())
	if err != nil {
		failure(writer, err)
		return
	}
	payload := make([]technicianResponse, 0, len(listed))
	for _, item := range listed {
		payload = append(payload, toTechnicianResponse(item))
	}
	respond(writer, http.StatusOK, payload)
}

// Create provisions a new technician employee account (atomic all-or-nothing).
func (h TechnicianHandler) Create(writer http.ResponseWriter, request *http.Request) {
	if _, err := requireAdministrator(request.Context()); err != nil {
		failure(writer, err)
		return
	}
	var payload createTechnicianRequest
	if err := decode(writer, request, &payload); err != nil {
		failure(writer, err)
		return
	}
	created, err := h.technician.Create(
		request.Context(),
		payload.FullName,
		payload.Username,
		payload.Password,
		payload.Specialty,
	)
	if err != nil {
		failure(writer, err)
		return
	}
	workload := domain.TechnicianWorkload{
		Technician: created,
		FullName:   payload.FullName,
		IsActive:   true,
		Busy:       false,
	}
	respond(writer, http.StatusCreated, toTechnicianResponse(workload))
}

// SetAccess modifies the access state of a technician account idempotently under canonical lock order.
func (h TechnicianHandler) SetAccess(writer http.ResponseWriter, request *http.Request) {
	caller, err := requireAdministrator(request.Context())
	if err != nil {
		failure(writer, err)
		return
	}
	technicianID := request.PathValue("technicianId")
	var payload setAccessRequest
	if err := decode(writer, request, &payload); err != nil {
		failure(writer, err)
		return
	}
	updated, err := h.technician.SetAccess(request.Context(), caller.UserID, technicianID, payload.Active)
	if err != nil {
		failure(writer, err)
		return
	}
	respond(writer, http.StatusOK, toTechnicianResponse(updated))
}

func toTechnicianResponse(item domain.TechnicianWorkload) technicianResponse {
	return technicianResponse{
		ID:                   item.Technician.ID,
		UserID:               item.Technician.UserID,
		FullName:             item.FullName,
		Specialty:            item.Technician.Specialty,
		IsActive:             item.IsActive,
		Busy:                 item.Busy,
		CanReceiveAssignment: item.CanReceiveAssignment(),
		ActiveOrderID:        item.ActiveOrderID,
		ActiveOrderNumber:    item.ActiveOrderNumber,
		ActiveVehiclePlate:   item.ActiveVehiclePlate,
	}
}
