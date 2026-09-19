package usecase

import (
	"context"
	"errors"
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
	ListByVehicle(ctx context.Context, vehicleID string) ([]domain.ServiceOrder, error)
	UpdateStatus(ctx context.Context, order domain.ServiceOrder, transition domain.StatusTransition) error
	ListTransition(ctx context.Context, serviceOrderID string) ([]domain.StatusTransition, error)
	CountByStatus(ctx context.Context) (map[string]int, error)
	NextOrderNumber(ctx context.Context) (string, error)
}

// ServiceOrderUseCase opens orders at check-in and advances their lifecycle.
type ServiceOrderUseCase struct {
	order      ServiceOrderRepository
	vehicle    VehicleRepository
	assignment AssignmentRepository
	technician TechnicianReader
	newID      func() string
	now        func() time.Time
}

// NewServiceOrderUseCase wires the service order use case.
func NewServiceOrderUseCase(
	order ServiceOrderRepository,
	vehicle VehicleRepository,
	assignment AssignmentRepository,
	technician TechnicianReader,
	newID func() string,
	now func() time.Time,
) ServiceOrderUseCase {
	return ServiceOrderUseCase{
		order:      order,
		vehicle:    vehicle,
		assignment: assignment,
		technician: technician,
		newID:      newID,
		now:        now,
	}
}

// Open registers the check-in of a vehicle and returns the created order.
func (s ServiceOrderUseCase) Open(ctx context.Context, vehicleID, reportedFailure string) (domain.ServiceOrder, error) {
	if _, err := s.vehicle.FindByID(ctx, vehicleID); err != nil {
		return domain.ServiceOrder{}, err
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

// List returns the orders, optionally filtered by a lifecycle status and scoped by actor role.
func (s ServiceOrderUseCase) List(ctx context.Context, status string, actorRole domain.Role, actorUserID string) ([]ServiceOrderSummary, error) {
	all, err := s.order.List(ctx, status)
	if err != nil {
		return nil, err
	}
	if actorRole != domain.RoleTechnician {
		return all, nil
	}
	actorTech, err := s.technician.FindByUserID(ctx, actorUserID)
	if err != nil {
		return []ServiceOrderSummary{}, nil
	}
	filtered := make([]ServiceOrderSummary, 0, len(all))
	for _, item := range all {
		if active, err := s.assignment.FindActiveByServiceOrder(ctx, item.Order.ID); err == nil {
			if active.TechnicianID == actorTech.ID {
				filtered = append(filtered, item)
				continue
			}
		}
		if latest, err := s.assignment.FindLatestByServiceOrder(ctx, item.Order.ID); err == nil {
			if latest.TechnicianID == actorTech.ID {
				filtered = append(filtered, item)
			}
		}
	}
	return filtered, nil
}

// Find returns one order by its identifier.
func (s ServiceOrderUseCase) Find(ctx context.Context, orderID string) (domain.ServiceOrder, error) {
	return s.order.FindByID(ctx, orderID)
}

// FindDetail returns the enriched detail of an order: the order entity, its vehicle,
// the assigned technician's name, permissions calculated for the actor, and whether the technician is active.
func (s ServiceOrderUseCase) FindDetail(
	ctx context.Context,
	orderID string,
	actorRole domain.Role,
	actorUserID string,
) (order domain.ServiceOrder, vehicle domain.Vehicle, technicianName string, permissions domain.OrderPermissions, technicianIsActive *bool, err error) {
	order, err = s.order.FindByID(ctx, orderID)
	if err != nil {
		return domain.ServiceOrder{}, domain.Vehicle{}, "", domain.OrderPermissions{}, nil, err
	}

	vehicleRecord, err := s.vehicle.FindByID(ctx, order.VehicleID)
	if err != nil {
		return domain.ServiceOrder{}, domain.Vehicle{}, "", domain.OrderPermissions{}, nil, err
	}
	vehicle = vehicleRecord.Vehicle

	var activeTechnicianID string
	var responsibleTechnicianID string
	if active, err := s.assignment.FindActiveByServiceOrder(ctx, orderID); err == nil {
		activeTechnicianID = active.TechnicianID
		responsibleTechnicianID = active.TechnicianID
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.ServiceOrder{}, domain.Vehicle{}, "", domain.OrderPermissions{}, nil, err
	} else if latest, err := s.assignment.FindLatestByServiceOrder(ctx, orderID); err == nil {
		responsibleTechnicianID = latest.TechnicianID
	}

	if responsibleTechnicianID != "" {
		if workloads, err := s.technician.ListWorkload(ctx); err == nil {
			for _, w := range workloads {
				if w.Technician.ID == responsibleTechnicianID {
					technicianName = w.FullName
					active := w.IsActive
					technicianIsActive = &active
					break
				}
			}
		}
	}

	var actorTechID string
	if actorRole == domain.RoleTechnician {
		actorTech, err := s.technician.FindByUserID(ctx, actorUserID)
		if err != nil {
			return domain.ServiceOrder{}, domain.Vehicle{}, "", domain.OrderPermissions{}, nil, domain.ErrForbidden
		}
		actorTechID = actorTech.ID
		if actorTechID == "" || (actorTechID != activeTechnicianID && actorTechID != responsibleTechnicianID) {
			return domain.ServiceOrder{}, domain.Vehicle{}, "", domain.OrderPermissions{}, nil, domain.ErrForbidden
		}
	}

	permissions = domain.CalculateOrderPermissions(actorRole, actorTechID, order, activeTechnicianID)
	return order, vehicle, technicianName, permissions, technicianIsActive, nil
}

// FindActiveAssignment returns the active technician assignment for a service order, if one exists.
func (s ServiceOrderUseCase) FindActiveAssignment(ctx context.Context, orderID string) (domain.Assignment, error) {
	return s.assignment.FindActiveByServiceOrder(ctx, orderID)
}

// ListTransition returns the status history of an order.
func (s ServiceOrderUseCase) ListTransition(ctx context.Context, orderID string) ([]domain.StatusTransition, error) {
	return s.order.ListTransition(ctx, orderID)
}

// Advance moves an order to the next status. The domain rejects a move outside
// the lifecycle before anything is written, so the stored order is untouched.
// Reaching DELIVERED releases the technician who held the order.
func (s ServiceOrderUseCase) Advance(ctx context.Context, orderID string, next domain.ServiceOrderStatus, actorRole domain.Role, actorUserID string) (domain.ServiceOrder, error) {
	order, err := s.order.FindByID(ctx, orderID)
	if err != nil {
		return domain.ServiceOrder{}, err
	}

	if actorRole != domain.RoleAdministrator {
		profile, err := s.technician.FindByUserID(ctx, actorUserID)
		if err != nil {
			return domain.ServiceOrder{}, domain.ErrForbidden
		}
		active, err := s.assignment.FindActiveByServiceOrder(ctx, orderID)
		if err != nil || active.TechnicianID != profile.ID {
			return domain.ServiceOrder{}, domain.ErrForbidden
		}
	}

	if next == domain.StatusInDiagnosis || next == domain.StatusInRepair {
		active, err := s.assignment.FindActiveByServiceOrder(ctx, orderID)
		if err != nil || active.TechnicianID == "" {
			return domain.ServiceOrder{}, domain.ErrTechnicianRequired
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
