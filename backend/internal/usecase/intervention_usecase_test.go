package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)



func TestInterventionRegisterRejectsForbiddenStatuses(t *testing.T) {
	tech, _ := domain.NewTechnician("tech-1", "user-tech-1", "Mecanica", fixedClock()())
	techRepo := newFakeTechnicianRepository(tech)

	forbiddenStatuses := []domain.ServiceOrderStatus{
		domain.StatusReceived,
		domain.StatusReady,
		domain.StatusDelivered,
	}

	for _, status := range forbiddenStatuses {
		order := buildOrder(t, "order-1", "OS-0001")
		order.Status = status
		orders := newFakeOrderRepository(order)

		assignments := &fakeAssignmentRepository{}
		assignment, _ := domain.NewAssignment("assign-1", "order-1", "tech-1", fixedClock()())
		_ = assignments.Save(context.Background(), assignment)

		uc := usecase.NewInterventionUseCase(
			newFakeInterventionRepository(t, "intervention-1"), orders, assignments, techRepo, sequentialID(), fixedClock(),
		)

		_, err := uc.Register(context.Background(), "order-1", "user-tech-1", "Reparacion", 2.0, nil)
		if err == nil {
			t.Fatalf("expected error when adding intervention to order in %s status", status)
		}
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict for status %s, got %v", status, err)
		}
		if !strings.Contains(err.Error(), "cannot add intervention to order in status") {
			t.Fatalf("expected error message mentioning cannot add intervention, got %v", err)
		}
	}
}

func TestInterventionRegisterSucceedsForInDiagnosis(t *testing.T) {
	tech, _ := domain.NewTechnician("tech-1", "user-tech-1", "Mecanica", fixedClock()())
	techRepo := newFakeTechnicianRepository(tech)

	order := buildOrder(t, "order-1", "OS-0001")
	order.Status = domain.StatusInDiagnosis
	orders := newFakeOrderRepository(order)

	assignments := &fakeAssignmentRepository{}
	assignment, _ := domain.NewAssignment("assign-1", "order-1", "tech-1", fixedClock()())
	_ = assignments.Save(context.Background(), assignment)

	uc := usecase.NewInterventionUseCase(
		newFakeInterventionRepository(t, "intervention-1"), orders, assignments, techRepo, sequentialID(), fixedClock(),
	)

	intervention, err := uc.Register(context.Background(), "order-1", "user-tech-1", "Cambio de aceite", 1.5, []usecase.PartUsageInput{
		{PartName: "Aceite sintético", Quantity: 4},
	})
	if err != nil {
		t.Fatalf("registering intervention in IN_DIAGNOSIS should succeed: %v", err)
	}
	if intervention.ServiceOrderID != "order-1" {
		t.Fatalf("unexpected order ID in intervention: %s", intervention.ServiceOrderID)
	}

	updatedOrder, _ := orders.FindByID(context.Background(), "order-1")
	if updatedOrder.Status != domain.StatusInRepair {
		t.Fatalf("order status should transition to IN_REPAIR, got %s", updatedOrder.Status)
	}
}
