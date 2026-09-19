package domain

import (
	"fmt"
	"strings"
	"time"
)

// ServiceOrderStatus is a step of the workshop service lifecycle.
type ServiceOrderStatus string

const (
	// StatusReceived is the status a service order is created with at check-in.
	StatusReceived ServiceOrderStatus = "RECEIVED"
	// StatusInDiagnosis means the assigned technician recorded the diagnostic.
	StatusInDiagnosis ServiceOrderStatus = "IN_DIAGNOSIS"
	// StatusInRepair means at least one intervention has been registered.
	StatusInRepair ServiceOrderStatus = "IN_REPAIR"
	// StatusReady means the vehicle is repaired and waiting for its owner.
	StatusReady ServiceOrderStatus = "READY"
	// StatusDelivered is the terminal status: the vehicle left the workshop.
	StatusDelivered ServiceOrderStatus = "DELIVERED"
)

// allowedTransition is the single source of truth for the lifecycle. Adding a
// status means editing this table and nothing else in the layers above.
var allowedTransition = map[ServiceOrderStatus][]ServiceOrderStatus{
	StatusReceived:    {StatusInDiagnosis},
	StatusInDiagnosis: {StatusInRepair},
	StatusInRepair:    {StatusReady},
	StatusReady:       {StatusDelivered},
	StatusDelivered:   {},
}

// Valid reports whether the status is one the lifecycle declares.
func (s ServiceOrderStatus) Valid() bool {
	_, known := allowedTransition[s]
	return known
}

// Next returns the next lifecycle status and reports whether one exists.
func (s ServiceOrderStatus) Next() (ServiceOrderStatus, bool) {
	candidates, ok := allowedTransition[s]
	if !ok || len(candidates) == 0 {
		return "", false
	}
	return candidates[0], true
}

// CanMoveTo reports whether the lifecycle allows moving to the next status.
func (s ServiceOrderStatus) CanMoveTo(next ServiceOrderStatus) bool {
	for _, candidate := range allowedTransition[s] {
		if candidate == next {
			return true
		}
	}
	return false
}

// IsOpen reports whether the order still occupies the workshop.
func (s ServiceOrderStatus) IsOpen() bool {
	return s != StatusDelivered
}

// CanAddDiagnostic reports whether a diagnostic evaluation can be attached to the order.
// Only allowed if status is RECEIVED or IN_DIAGNOSIS.
func (s ServiceOrderStatus) CanAddDiagnostic() bool {
	return s == StatusReceived || s == StatusInDiagnosis
}

// CanAddIntervention reports whether a repair intervention can be recorded.
// Only allowed if status is IN_DIAGNOSIS or IN_REPAIR.
func (s ServiceOrderStatus) CanAddIntervention() bool {
	return s == StatusInDiagnosis || s == StatusInRepair
}

// CanAssignTechnician reports whether a technician can be assigned or reassigned.
// Only allowed if status is not DELIVERED.
func (s ServiceOrderStatus) CanAssignTechnician() bool {
	return s != StatusDelivered
}

// IsOperational reports whether the order is currently being actively worked on.
func (s ServiceOrderStatus) IsOperational() bool {
	return s == StatusInDiagnosis || s == StatusInRepair
}

// ServiceOrder is the work order opened at check-in for one vehicle.
type ServiceOrder struct {
	ID              string
	OrderNumber     string
	VehicleID       string
	ReportedFailure string
	Status          ServiceOrderStatus
	ReceivedAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// StatusTransition is the audit record of one status change. It is always
// written in the same transaction as the change it describes.
type StatusTransition struct {
	ID              string
	ServiceOrderID  string
	FromStatus      ServiceOrderStatus
	ToStatus        ServiceOrderStatus
	ChangedByUserID string
	ChangedByName   string
	ChangedAt       time.Time
}

// NewServiceOrder opens a service order in RECEIVED status.
func NewServiceOrder(id, orderNumber, vehicleID, reportedFailure string, receivedAt time.Time) (ServiceOrder, error) {
	orderNumber = strings.TrimSpace(orderNumber)
	reportedFailure = strings.TrimSpace(reportedFailure)
	if id == "" {
		return ServiceOrder{}, fmt.Errorf("%w: service order identifier is required", ErrInvalidInput)
	}
	if orderNumber == "" {
		return ServiceOrder{}, fmt.Errorf("%w: service order number is required", ErrInvalidInput)
	}
	if vehicleID == "" {
		return ServiceOrder{}, fmt.Errorf("%w: the vehicle of the service order is required", ErrInvalidInput)
	}
	if reportedFailure == "" {
		return ServiceOrder{}, fmt.Errorf("%w: the reported failure is required", ErrInvalidInput)
	}
	return ServiceOrder{
		ID:              id,
		OrderNumber:     orderNumber,
		VehicleID:       vehicleID,
		ReportedFailure: reportedFailure,
		Status:          StatusReceived,
		ReceivedAt:      receivedAt,
		CreatedAt:       receivedAt,
		UpdatedAt:       receivedAt,
	}, nil
}

// MoveTo advances the order and returns the transition record to persist with
// it. A move outside the lifecycle leaves the order exactly as it was.
func (o *ServiceOrder) MoveTo(next ServiceOrderStatus, transitionID, changedByUserID string, at time.Time) (StatusTransition, error) {
	if !next.Valid() {
		return StatusTransition{}, fmt.Errorf("%w: unknown status %q", ErrInvalidTransition, next)
	}
	if !o.Status.CanMoveTo(next) {
		return StatusTransition{}, fmt.Errorf("%w: %s cannot move to %s", ErrInvalidTransition, o.Status, next)
	}
	if transitionID == "" || changedByUserID == "" {
		return StatusTransition{}, fmt.Errorf("%w: the transition author is required", ErrInvalidInput)
	}
	transition := StatusTransition{
		ID:              transitionID,
		ServiceOrderID:  o.ID,
		FromStatus:      o.Status,
		ToStatus:        next,
		ChangedByUserID: changedByUserID,
		ChangedAt:       at,
	}
	o.Status = next
	o.UpdatedAt = at
	return transition, nil
}

// CanAddDiagnostic reports whether a diagnostic evaluation can be attached to the order.
func (o ServiceOrder) CanAddDiagnostic() bool {
	return o.Status.CanAddDiagnostic()
}

// CanAddIntervention reports whether a repair intervention can be recorded.
func (o ServiceOrder) CanAddIntervention() bool {
	return o.Status.CanAddIntervention()
}

// CanAssignTechnician reports whether a technician can be assigned or reassigned.
func (o ServiceOrder) CanAssignTechnician() bool {
	return o.Status.CanAssignTechnician()
}

// IsOperational reports whether the order is currently being actively worked on.
func (o ServiceOrder) IsOperational() bool {
	return o.Status.IsOperational()
}
