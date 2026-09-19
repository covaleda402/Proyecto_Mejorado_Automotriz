package domain_test

import (
	"testing"
	"time"

	"workshop/internal/domain"
)

func makeTestOrder(status domain.ServiceOrderStatus) domain.ServiceOrder {
	return domain.ServiceOrder{
		ID:              "order-1",
		OrderNumber:     "OS-0001",
		VehicleID:       "vehicle-1",
		ReportedFailure: "Check engine light",
		Status:          status,
		ReceivedAt:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func TestStatusLifecyclePolicies(t *testing.T) {
	cases := []struct {
		status          domain.ServiceOrderStatus
		canDiagnostic   bool
		canIntervention bool
		canAssign       bool
		isOperational   bool
	}{
		{domain.StatusReceived, true, false, true, false},
		{domain.StatusInDiagnosis, true, true, true, true},
		{domain.StatusInRepair, false, true, true, true},
		{domain.StatusReady, false, false, true, false},
		{domain.StatusDelivered, false, false, false, false},
	}

	for _, tc := range cases {
		order := makeTestOrder(tc.status)

		if tc.status.CanAddDiagnostic() != tc.canDiagnostic {
			t.Errorf("status %s CanAddDiagnostic = %v, expected %v", tc.status, tc.status.CanAddDiagnostic(), tc.canDiagnostic)
		}
		if order.CanAddDiagnostic() != tc.canDiagnostic {
			t.Errorf("order %s CanAddDiagnostic = %v, expected %v", order.Status, order.CanAddDiagnostic(), tc.canDiagnostic)
		}

		if tc.status.CanAddIntervention() != tc.canIntervention {
			t.Errorf("status %s CanAddIntervention = %v, expected %v", tc.status, tc.status.CanAddIntervention(), tc.canIntervention)
		}
		if order.CanAddIntervention() != tc.canIntervention {
			t.Errorf("order %s CanAddIntervention = %v, expected %v", order.Status, order.CanAddIntervention(), tc.canIntervention)
		}

		if tc.status.CanAssignTechnician() != tc.canAssign {
			t.Errorf("status %s CanAssignTechnician = %v, expected %v", tc.status, tc.status.CanAssignTechnician(), tc.canAssign)
		}
		if order.CanAssignTechnician() != tc.canAssign {
			t.Errorf("order %s CanAssignTechnician = %v, expected %v", order.Status, order.CanAssignTechnician(), tc.canAssign)
		}

		if tc.status.IsOperational() != tc.isOperational {
			t.Errorf("status %s IsOperational = %v, expected %v", tc.status, tc.status.IsOperational(), tc.isOperational)
		}
		if order.IsOperational() != tc.isOperational {
			t.Errorf("order %s IsOperational = %v, expected %v", order.Status, order.IsOperational(), tc.isOperational)
		}
	}
}

func TestCalculateOrderPermissionsForAdministrator(t *testing.T) {
	// Administrator on open statuses: CanAdvance=true, CanAssign=true, CanAddDiagnostic=false, CanAddIntervention=false
	openStatuses := []domain.ServiceOrderStatus{
		domain.StatusReceived,
		domain.StatusInDiagnosis,
		domain.StatusInRepair,
		domain.StatusReady,
	}

	for _, s := range openStatuses {
		order := makeTestOrder(s)
		perms := domain.CalculateOrderPermissions(domain.RoleAdministrator, "", order, "tech-1")
		if !perms.CanAdvance {
			t.Errorf("admin on %s should have CanAdvance = true", s)
		}
		if !perms.CanAssign {
			t.Errorf("admin on %s should have CanAssign = true", s)
		}
		if perms.CanAddDiagnostic {
			t.Errorf("admin on %s should have CanAddDiagnostic = false", s)
		}
		if perms.CanAddIntervention {
			t.Errorf("admin on %s should have CanAddIntervention = false", s)
		}
	}

	// Delivered order: all false
	deliveredOrder := makeTestOrder(domain.StatusDelivered)
	deliveredPerms := domain.CalculateOrderPermissions(domain.RoleAdministrator, "", deliveredOrder, "tech-1")
	if deliveredPerms != (domain.OrderPermissions{}) {
		t.Errorf("admin on DELIVERED should have all permissions false, got %+v", deliveredPerms)
	}
}

func TestCalculateOrderPermissionsForTechnician(t *testing.T) {
	techID := "tech-1"
	otherTechID := "tech-2"

	// Non-matching technician gets all false
	order := makeTestOrder(domain.StatusInDiagnosis)
	permsNonMatch := domain.CalculateOrderPermissions(domain.RoleTechnician, techID, order, otherTechID)
	if permsNonMatch != (domain.OrderPermissions{}) {
		t.Errorf("non-matching technician should have all false, got %+v", permsNonMatch)
	}

	// Empty technician ID gets all false
	permsEmpty := domain.CalculateOrderPermissions(domain.RoleTechnician, "", order, "")
	if permsEmpty != (domain.OrderPermissions{}) {
		t.Errorf("empty technician ID should have all false, got %+v", permsEmpty)
	}

	// Matching technician on RECEIVED
	receivedOrder := makeTestOrder(domain.StatusReceived)
	permsReceived := domain.CalculateOrderPermissions(domain.RoleTechnician, techID, receivedOrder, techID)
	if !permsReceived.CanAdvance || !permsReceived.CanAddDiagnostic || permsReceived.CanAddIntervention || permsReceived.CanAssign {
		t.Errorf("tech on RECEIVED permissions mismatch: %+v", permsReceived)
	}

	// Matching technician on IN_DIAGNOSIS
	diagOrder := makeTestOrder(domain.StatusInDiagnosis)
	permsDiag := domain.CalculateOrderPermissions(domain.RoleTechnician, techID, diagOrder, techID)
	if !permsDiag.CanAdvance || !permsDiag.CanAddDiagnostic || !permsDiag.CanAddIntervention || permsDiag.CanAssign {
		t.Errorf("tech on IN_DIAGNOSIS permissions mismatch: %+v", permsDiag)
	}

	// Matching technician on IN_REPAIR
	repairOrder := makeTestOrder(domain.StatusInRepair)
	permsRepair := domain.CalculateOrderPermissions(domain.RoleTechnician, techID, repairOrder, techID)
	if !permsRepair.CanAdvance || permsRepair.CanAddDiagnostic || !permsRepair.CanAddIntervention || permsRepair.CanAssign {
		t.Errorf("tech on IN_REPAIR permissions mismatch: %+v", permsRepair)
	}

	// Matching technician on READY
	readyOrder := makeTestOrder(domain.StatusReady)
	permsReady := domain.CalculateOrderPermissions(domain.RoleTechnician, techID, readyOrder, techID)
	if !permsReady.CanAdvance || permsReady.CanAddDiagnostic || permsReady.CanAddIntervention || permsReady.CanAssign {
		t.Errorf("tech on READY permissions mismatch: %+v", permsReady)
	}

	// Delivered order: all false
	deliveredOrder := makeTestOrder(domain.StatusDelivered)
	permsDelivered := domain.CalculateOrderPermissions(domain.RoleTechnician, techID, deliveredOrder, techID)
	if permsDelivered != (domain.OrderPermissions{}) {
		t.Errorf("tech on DELIVERED should have all false, got %+v", permsDelivered)
	}
}

func TestCalculateOrderPermissionsUnknownRole(t *testing.T) {
	order := makeTestOrder(domain.StatusInDiagnosis)
	perms := domain.CalculateOrderPermissions("UNKNOWN_ROLE", "tech-1", order, "tech-1")
	if perms != (domain.OrderPermissions{}) {
		t.Errorf("unknown role should have all false, got %+v", perms)
	}
}
