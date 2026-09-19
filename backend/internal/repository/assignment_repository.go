package repository

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/domain"
)

const assignmentColumn = "id, service_order_id, technician_id, is_active, active_marker, active_order_marker, assigned_at, released_at"

// AssignmentRepository persists which technician holds which service order.
type AssignmentRepository struct {
	database *sql.DB
	timeout  time.Duration
}

// NewAssignmentRepository wires the assignment adapter.
func NewAssignmentRepository(database *sql.DB, timeout time.Duration) AssignmentRepository {
	return AssignmentRepository{database: database, timeout: timeout}
}

// Save stores an assignment. The unique active markers turn a second active
// assignment for the same technician or same service order into a conflict, even under a race.
func (r AssignmentRepository) Save(ctx context.Context, assignment domain.Assignment) error {
	return r.AssignWithLock(ctx, assignment)
}

// AssignWithLock stores an assignment within an exclusive transaction following the canonical lock order:
// 1. Lock user row associated with technician FOR UPDATE
// 2. Lock technician row FOR UPDATE
// 3. Verify that user.is_active == 1 (if not, returns domain.ErrTechnicianInactive)
// 4. Insert assignment row (physically guarded by uq_assignment_active_marker and uq_assignment_active_order_marker)
func (r AssignmentRepository) AssignWithLock(ctx context.Context, assignment domain.Assignment) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	tx, err := r.database.BeginTx(queryCtx, nil)
	if err != nil {
		return translate(err)
	}
	defer func() { _ = tx.Rollback() }()

	// Canonical Lock Step 1: Lock user row associated with technician
	var userID string
	var isActive int
	err = tx.QueryRowContext(queryCtx,
		"SELECT u.id, u.is_active FROM `user` u JOIN technician t ON t.user_id = u.id WHERE t.id = ? FOR UPDATE",
		assignment.TechnicianID,
	).Scan(&userID, &isActive)
	if err != nil {
		return translate(err)
	}

	// Canonical Lock Step 2: Lock technician row
	var techID string
	err = tx.QueryRowContext(queryCtx,
		"SELECT id FROM technician WHERE id = ? FOR UPDATE",
		assignment.TechnicianID,
	).Scan(&techID)
	if err != nil {
		return translate(err)
	}

	// Invariant BR-ACTIVE-05: Technician must be active
	if isActive != 1 {
		return domain.ErrTechnicianInactive
	}

	// Step 4: Insert assignment (physically guarded by uq_assignment_active_marker and uq_assignment_active_order_marker)
	_, err = tx.ExecContext(queryCtx,
		"INSERT INTO assignment ("+assignmentColumn+") VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		assignment.ID, assignment.ServiceOrderID, assignment.TechnicianID,
		assignment.IsActive, assignment.ActiveMarker, assignment.ActiveOrderMarker, assignment.AssignedAt, assignment.ReleasedAt,
	)
	if err != nil {
		return translate(err)
	}

	return translate(tx.Commit())
}

// FindActiveByServiceOrder reads the technician currently holding an order.
func (r AssignmentRepository) FindActiveByServiceOrder(ctx context.Context, serviceOrderID string) (domain.Assignment, error) {
	return r.findActive(ctx, "service_order_id", serviceOrderID)
}

// FindActiveByTechnician reads the order a technician is currently holding.
func (r AssignmentRepository) FindActiveByTechnician(ctx context.Context, technicianID string) (domain.Assignment, error) {
	return r.findActive(ctx, "technician_id", technicianID)
}

// FindLatestByServiceOrder reads the most recent assignment for an order, active or released.
func (r AssignmentRepository) FindLatestByServiceOrder(ctx context.Context, serviceOrderID string) (domain.Assignment, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var assignment domain.Assignment
	var marker sql.NullString
	var orderMarker sql.NullString
	var releasedAt sql.NullTime
	err := r.database.QueryRowContext(
		queryCtx,
		"SELECT "+assignmentColumn+" FROM assignment WHERE service_order_id = ? ORDER BY assigned_at DESC LIMIT 1",
		serviceOrderID,
	).Scan(
		&assignment.ID, &assignment.ServiceOrderID, &assignment.TechnicianID,
		&assignment.IsActive, &marker, &orderMarker, &assignment.AssignedAt, &releasedAt,
	)
	if err != nil {
		return domain.Assignment{}, translate(err)
	}
	if marker.Valid {
		value := marker.String
		assignment.ActiveMarker = &value
	}
	if orderMarker.Valid {
		value := orderMarker.String
		assignment.ActiveOrderMarker = &value
	}
	if releasedAt.Valid {
		moment := releasedAt.Time
		assignment.ReleasedAt = &moment
	}
	return assignment, nil
}

// ReleaseByServiceOrder frees the technician when the vehicle is delivered.
func (r AssignmentRepository) ReleaseByServiceOrder(ctx context.Context, serviceOrderID string, releasedAt time.Time) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.database.ExecContext(
		queryCtx,
		"UPDATE assignment SET is_active = 0, active_marker = NULL, active_order_marker = NULL, released_at = ? "+
			"WHERE service_order_id = ? AND is_active = 1",
		releasedAt, serviceOrderID,
	)
	return translate(err)
}

func (r AssignmentRepository) findActive(ctx context.Context, column, value string) (domain.Assignment, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var assignment domain.Assignment
	var marker sql.NullString
	var orderMarker sql.NullString
	var releasedAt sql.NullTime
	err := r.database.QueryRowContext(
		queryCtx,
		"SELECT "+assignmentColumn+" FROM assignment WHERE "+column+" = ? AND is_active = 1",
		value,
	).Scan(
		&assignment.ID, &assignment.ServiceOrderID, &assignment.TechnicianID,
		&assignment.IsActive, &marker, &orderMarker, &assignment.AssignedAt, &releasedAt,
	)
	if err != nil {
		return domain.Assignment{}, translate(err)
	}
	if marker.Valid {
		value := marker.String
		assignment.ActiveMarker = &value
	}
	if orderMarker.Valid {
		value := orderMarker.String
		assignment.ActiveOrderMarker = &value
	}
	if releasedAt.Valid {
		moment := releasedAt.Time
		assignment.ReleasedAt = &moment
	}
	return assignment, nil
}
