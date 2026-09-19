package http

import (
	"net/http"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

// createServiceOrderRequest is the payload of the check-in form.
type createServiceOrderRequest struct {
	VehicleID       string `json:"vehicleId"`
	ReportedFailure string `json:"reportedFailure"`
}

// advanceStatusRequest is the payload of the status advance buttons.
type advanceStatusRequest struct {
	Status string `json:"status"`
}

// serviceOrderResponse is one row of the order table and the order detail.
type serviceOrderResponse struct {
	ID                   string                   `json:"id"`
	OrderNumber          string                   `json:"orderNumber"`
	VehicleID            string                   `json:"vehicleId"`
	VehiclePlate         string                   `json:"vehiclePlate"`
	TechnicianName       string                   `json:"technicianName"`
	AssignedTechnicianID string                   `json:"assignedTechnicianId,omitempty"`
	TechnicianIsActive   *bool                    `json:"technicianIsActive,omitempty"`
	ReportedFailure      string                   `json:"reportedFailure"`
	Status               string                   `json:"status"`
	ReceivedAt           string                   `json:"receivedAt"`
	UpdatedAt            string                   `json:"updatedAt"`
	Permissions          *domain.OrderPermissions `json:"permissions,omitempty"`
}

// statusTransitionResponse is one row of the status history panel.
type statusTransitionResponse struct {
	ID              string `json:"id"`
	FromStatus      string `json:"fromStatus"`
	ToStatus        string `json:"toStatus"`
	ChangedByUserID string `json:"changedByUserId"`
	ChangedByName   string `json:"changedByName"`
	ChangedAt       string `json:"changedAt"`
}

// ServiceOrderHandler exposes the check-in and the lifecycle of an order.
type ServiceOrderHandler struct {
	order usecase.ServiceOrderUseCase
}

// NewServiceOrderHandler wires the service order handler.
func NewServiceOrderHandler(order usecase.ServiceOrderUseCase) ServiceOrderHandler {
	return ServiceOrderHandler{order: order}
}

// Create opens a service order at check-in.
func (h ServiceOrderHandler) Create(writer http.ResponseWriter, request *http.Request) {
	if _, err := requireAdministrator(request.Context()); err != nil {
		failure(writer, err)
		return
	}
	var payload createServiceOrderRequest
	if err := decode(writer, request, &payload); err != nil {
		failure(writer, err)
		return
	}
	created, err := h.order.Open(request.Context(), payload.VehicleID, payload.ReportedFailure)
	if err != nil {
		failure(writer, err)
		return
	}
	respond(writer, http.StatusCreated, toServiceOrderResponse(created, "", ""))
}

// List returns the orders, optionally filtered by status and caller identity.
func (h ServiceOrderHandler) List(writer http.ResponseWriter, request *http.Request) {
	identity, err := callerFrom(request.Context())
	if err != nil {
		failure(writer, err)
		return
	}
	listed, err := h.order.List(request.Context(), request.URL.Query().Get("status"), identity.Role, identity.UserID)
	if err != nil {
		failure(writer, err)
		return
	}
	payload := make([]serviceOrderResponse, 0, len(listed))
	for _, item := range listed {
		payload = append(payload, toServiceOrderResponse(item.Order, item.VehiclePlate, item.TechnicianName))
	}
	respond(writer, http.StatusOK, payload)
}

// Find returns one order by its identifier.
func (h ServiceOrderHandler) Find(writer http.ResponseWriter, request *http.Request) {
	identity, err := callerFrom(request.Context())
	if err != nil {
		failure(writer, err)
		return
	}
	order, vehicle, technicianName, permissions, techIsActive, err := h.order.FindDetail(
		request.Context(), request.PathValue("serviceOrderId"), identity.Role, identity.UserID,
	)
	if err != nil {
		failure(writer, err)
		return
	}
	resp := toServiceOrderResponse(order, vehicle.Plate, technicianName)
	resp.Permissions = &permissions
	resp.TechnicianIsActive = techIsActive
	if active, err := h.order.FindActiveAssignment(request.Context(), order.ID); err == nil {
		resp.AssignedTechnicianID = active.TechnicianID
	}
	respond(writer, http.StatusOK, resp)
}

// ListTransition returns the status history of an order.
func (h ServiceOrderHandler) ListTransition(writer http.ResponseWriter, request *http.Request) {
	history, err := h.order.ListTransition(request.Context(), request.PathValue("serviceOrderId"))
	if err != nil {
		failure(writer, err)
		return
	}
	payload := make([]statusTransitionResponse, 0, len(history))
	for _, item := range history {
		payload = append(payload, statusTransitionResponse{
			ID:              item.ID,
			FromStatus:      string(item.FromStatus),
			ToStatus:        string(item.ToStatus),
			ChangedByUserID: item.ChangedByUserID,
			ChangedByName:   item.ChangedByName,
			ChangedAt:       formatTime(item.ChangedAt),
		})
	}
	respond(writer, http.StatusOK, payload)
}

// Advance moves an order to the next lifecycle status.
func (h ServiceOrderHandler) Advance(writer http.ResponseWriter, request *http.Request) {
	identity, err := callerFrom(request.Context())
	if err != nil {
		failure(writer, err)
		return
	}
	var payload advanceStatusRequest
	if err := decode(writer, request, &payload); err != nil {
		failure(writer, err)
		return
	}
	order, err := h.order.Advance(
		request.Context(),
		request.PathValue("serviceOrderId"),
		domain.ServiceOrderStatus(payload.Status),
		identity.Role,
		identity.UserID,
	)
	if err != nil {
		failure(writer, err)
		return
	}
	respond(writer, http.StatusOK, toServiceOrderResponse(order, "", ""))
}

func toServiceOrderResponse(order domain.ServiceOrder, plate, technicianName string) serviceOrderResponse {
	return serviceOrderResponse{
		ID:              order.ID,
		OrderNumber:     order.OrderNumber,
		VehicleID:       order.VehicleID,
		VehiclePlate:    plate,
		TechnicianName:  technicianName,
		ReportedFailure: order.ReportedFailure,
		Status:          string(order.Status),
		ReceivedAt:      formatTime(order.ReceivedAt),
		UpdatedAt:       formatTime(order.UpdatedAt),
	}
}
