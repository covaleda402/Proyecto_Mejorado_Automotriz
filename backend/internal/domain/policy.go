package domain

// OrderPermissions describes the allowed operations for a specific user on a service order.
type OrderPermissions struct {
	CanAdvance         bool `json:"canAdvance"`
	CanAddDiagnostic   bool `json:"canAddDiagnostic"`
	CanAddIntervention bool `json:"canAddIntervention"`
	CanAssign          bool `json:"canAssign"`
}

// CalculateOrderPermissions is a pure function evaluating authorization and lifecycle
// permissions based on the actor identity (role, technician ID), active order assignment,
// and current order status.
func CalculateOrderPermissions(
	actorRole Role,
	actorTechnicianID string,
	order ServiceOrder,
	activeTechnicianID string,
) OrderPermissions {
	if !order.Status.IsOpen() {
		return OrderPermissions{}
	}

	nextStatus, hasNext := order.Status.Next()
	canAdvance := hasNext && order.Status.CanMoveTo(nextStatus)
	// Invariant: cannot advance to an operational status (IN_DIAGNOSIS or IN_REPAIR) without an assigned technician.
	if (nextStatus == StatusInDiagnosis || nextStatus == StatusInRepair) && activeTechnicianID == "" {
		canAdvance = false
	}

	switch actorRole {
	case RoleAdministrator:
		return OrderPermissions{
			CanAdvance:         canAdvance,
			CanAddDiagnostic:   false,
			CanAddIntervention: false,
			CanAssign:          order.Status.CanAssignTechnician(),
		}
	case RoleTechnician:
		if actorTechnicianID == "" || actorTechnicianID != activeTechnicianID {
			return OrderPermissions{}
		}
		return OrderPermissions{
			CanAdvance:         canAdvance,
			CanAddDiagnostic:   order.Status.CanAddDiagnostic(),
			CanAddIntervention: order.Status.CanAddIntervention(),
			CanAssign:          false,
		}
	default:
		return OrderPermissions{}
	}
}
