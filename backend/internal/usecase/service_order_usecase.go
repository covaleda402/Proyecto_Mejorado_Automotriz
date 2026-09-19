package usecase

import (
	"context"
	"time"

	"workshop/internal/domain"
)

// ServiceOrderSummary is the read model of the order list: the order plus the
// plate of its vehicle and the name of the technician holding it.
type ServiceOrderSummary struct {
	Order          domain.ServiceOrder
	VehiclePlate   string
	TechnicianName string
}

// ServiceOrderRepository is the narrow port the service order use case needs.
// UpdateStatus writes the order and its transition record in one transaction.
type ServiceOrderRepository interface {
	Save(ctx context.Context, order domain.ServiceOrder) error
	FindByID(ctx context.Context, id string) (domain.ServiceOrder, error)
	List(ctx context.Context, status string) ([]ServiceOrderSummary, error)
	ListByTechnicianUser(ctx context.Context, userID, status string) ([]ServiceOrderSummary, error)
	ListByVehicle(ctx context.Context, vehicleID string) ([]domain.ServiceOrder, error)
	UpdateStatus(ctx context.Context, order domain.ServiceOrder, transition domain.StatusTransition) error
	ListTransition(ctx context.Context, serviceOrderID string) ([]domain.StatusTransition, error)
	CountByStatus(ctx context.Context) (map[string]int, error)
	NextOrderNumber(ctx context.Context) (string, error)
}

// ServiceOrderUseCase opens orders at check-in and advances their lifecycle.
type ServiceOrderUseCase struct {
	order        ServiceOrderRepository
	vehicle      VehicleRepository
	assignment   AssignmentRepository
	technician   TechnicianRepository
	diagnostic   DiagnosticRepository
	intervention InterventionRepository
	newID        func() string
	now          func() time.Time
}

// NewServiceOrderUseCase wires the service order use case.
func NewServiceOrderUseCase(
	order ServiceOrderRepository,
	vehicle VehicleRepository,
	assignment AssignmentRepository,
	technician TechnicianRepository,
	diagnostic DiagnosticRepository,
	intervention InterventionRepository,
	newID func() string,
	now func() time.Time,
) ServiceOrderUseCase {
	return ServiceOrderUseCase{
		order:        order,
		vehicle:      vehicle,
		assignment:   assignment,
		technician:   technician,
		diagnostic:   diagnostic,
		intervention: intervention,
		newID:        newID,
		now:          now,
	}
}

// Open registers the check-in of a vehicle and returns the created order.
// A vehicle can have multiple historic orders, but cannot have two active orders.
func (s ServiceOrderUseCase) Open(ctx context.Context, vehicleID, reportedFailure string) (domain.ServiceOrder, error) {
	if _, err := s.vehicle.FindByID(ctx, vehicleID); err != nil {
		return domain.ServiceOrder{}, err
	}
	existingOrders, err := s.order.ListByVehicle(ctx, vehicleID)
	if err == nil {
		for _, existing := range existingOrders {
			if existing.Status.IsOpen() {
				return domain.ServiceOrder{}, domain.ErrConflict
			}
		}
	}
	orderNumber, err := s.order.NextOrderNumber(ctx)
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	order, err := domain.NewServiceOrder(s.newID(), orderNumber, vehicleID, reportedFailure, s.now())
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	if err := s.order.Save(ctx, order); err != nil {
		return domain.ServiceOrder{}, err
	}
	return order, nil
}

// List returns the orders, optionally filtered by a lifecycle status and scoped to caller.
func (s ServiceOrderUseCase) List(ctx context.Context, status, callerUserID string, callerRole domain.Role) ([]ServiceOrderSummary, error) {
	if callerRole == domain.RoleAdministrator {
		return s.order.List(ctx, status)
	}
	return s.order.ListByTechnicianUser(ctx, callerUserID, status)
}

// Find returns one order by its identifier, checking technician access.
func (s ServiceOrderUseCase) Find(ctx context.Context, orderID, callerUserID string, callerRole domain.Role) (domain.ServiceOrder, error) {
	order, err := s.order.FindByID(ctx, orderID)
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	if callerRole == domain.RoleTechnician {
		if s.technician != nil && s.assignment != nil {
			profile, err := s.technician.FindByUserID(ctx, callerUserID)
			if err != nil {
				return domain.ServiceOrder{}, domain.ErrForbidden
			}
			active, err := s.assignment.FindActiveByServiceOrder(ctx, orderID)
			if err != nil || active.TechnicianID != profile.ID {
				return domain.ServiceOrder{}, domain.ErrForbidden
			}
		}
	}
	return order, nil
}

// ListTransition returns the status history of an order.
func (s ServiceOrderUseCase) ListTransition(ctx context.Context, orderID string) ([]domain.StatusTransition, error) {
	return s.order.ListTransition(ctx, orderID)
}

// Advance moves an order to the next status, enforcing flow and ownership rules.
func (s ServiceOrderUseCase) Advance(
	ctx context.Context,
	orderID string,
	next domain.ServiceOrderStatus,
	actorUserID string,
	actorRole domain.Role,
) (domain.ServiceOrder, error) {
	order, err := s.order.FindByID(ctx, orderID)
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	if order.Status == domain.StatusDelivered {
		return domain.ServiceOrder{}, domain.ErrForbidden
	}
	if actorRole == domain.RoleTechnician {
		if s.technician != nil && s.assignment != nil {
			profile, err := s.technician.FindByUserID(ctx, actorUserID)
			if err != nil {
				return domain.ServiceOrder{}, domain.ErrForbidden
			}
			active, err := s.assignment.FindActiveByServiceOrder(ctx, orderID)
			if err != nil || active.TechnicianID != profile.ID {
				return domain.ServiceOrder{}, domain.ErrForbidden
			}
		}
	}

	// Requirements before moving status:
	if next == domain.StatusInDiagnosis && s.diagnostic != nil {
		if _, err := s.diagnostic.FindByServiceOrder(ctx, orderID); err != nil {
			return domain.ServiceOrder{}, domain.ErrInvalidTransition
		}
	}
	if next == domain.StatusInRepair && s.intervention != nil {
		interventions, err := s.intervention.ListByServiceOrder(ctx, orderID)
		if err != nil || len(interventions) == 0 {
			return domain.ServiceOrder{}, domain.ErrInvalidTransition
		}
	}

	changedAt := s.now()
	transition, err := order.MoveTo(next, s.newID(), actorUserID, changedAt)
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	if err := s.order.UpdateStatus(ctx, order, transition); err != nil {
		return domain.ServiceOrder{}, err
	}
	if next == domain.StatusDelivered {
		if err := s.assignment.ReleaseByServiceOrder(ctx, order.ID, changedAt); err != nil {
			return domain.ServiceOrder{}, err
		}
	}
	return order, nil
}
