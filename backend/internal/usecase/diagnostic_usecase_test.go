package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type fakeDiagnosticRepository struct {
	diagnostic map[string]domain.Diagnostic
}

func newFakeDiagnosticRepository() *fakeDiagnosticRepository {
	return &fakeDiagnosticRepository{diagnostic: make(map[string]domain.Diagnostic)}
}

func (f *fakeDiagnosticRepository) Save(_ context.Context, diagnostic domain.Diagnostic) error {
	f.diagnostic[diagnostic.ID] = diagnostic
	return nil
}

func (f *fakeDiagnosticRepository) FindByServiceOrder(_ context.Context, serviceOrderID string) (domain.Diagnostic, error) {
	for _, d := range f.diagnostic {
		if d.ServiceOrderID == serviceOrderID {
			return d, nil
		}
	}
	return domain.Diagnostic{}, domain.ErrNotFound
}

func (f *fakeDiagnosticRepository) ListByVehicle(_ context.Context, _ string) ([]domain.Diagnostic, error) {
	return nil, nil
}

func TestDiagnosticRecordRejectsOrderInReadyOrDeliveredStatus(t *testing.T) {
	tech, _ := domain.NewTechnician("tech-1", "user-tech-1", "Mecanica", fixedClock()())
	techRepo := newFakeTechnicianRepository(tech)

	for _, status := range []domain.ServiceOrderStatus{domain.StatusReady, domain.StatusDelivered} {
		order := buildOrder(t, "order-1", "OS-0001")
		order.Status = status
		orders := newFakeOrderRepository(order)

		assignments := &fakeAssignmentRepository{}
		assignment, _ := domain.NewAssignment("assign-1", "order-1", "tech-1", fixedClock()())
		_ = assignments.Save(context.Background(), assignment)

		uc := usecase.NewDiagnosticUseCase(
			newFakeDiagnosticRepository(), orders, assignments, techRepo, sequentialID(), fixedClock(),
		)

		_, err := uc.Record(context.Background(), "order-1", "user-tech-1", "Falla detectada", "Motor")
		if err == nil {
			t.Fatalf("expected error when adding diagnostic to order in %s status", status)
		}
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict, got %v", err)
		}
		if !strings.Contains(err.Error(), "cannot add diagnostic to order in status") {
			t.Fatalf("expected error message mentioning cannot add diagnostic, got %v", err)
		}
	}
}

func TestDiagnosticRecordSucceedsForReceivedOrder(t *testing.T) {
	tech, _ := domain.NewTechnician("tech-1", "user-tech-1", "Mecanica", fixedClock()())
	techRepo := newFakeTechnicianRepository(tech)
	order := buildOrder(t, "order-1", "OS-0001")
	orders := newFakeOrderRepository(order)

	assignments := &fakeAssignmentRepository{}
	assignment, _ := domain.NewAssignment("assign-1", "order-1", "tech-1", fixedClock()())
	_ = assignments.Save(context.Background(), assignment)

	diagRepo := newFakeDiagnosticRepository()
	uc := usecase.NewDiagnosticUseCase(
		diagRepo, orders, assignments, techRepo, sequentialID(), fixedClock(),
	)

	diagnostic, err := uc.Record(context.Background(), "order-1", "user-tech-1", "Bujias danadas", "Bujias")
	if err != nil {
		t.Fatalf("recording diagnostic for RECEIVED order should succeed: %v", err)
	}
	if diagnostic.ServiceOrderID != "order-1" {
		t.Fatalf("unexpected service order ID: %s", diagnostic.ServiceOrderID)
	}

	updatedOrder, _ := orders.FindByID(context.Background(), "order-1")
	if updatedOrder.Status != domain.StatusInDiagnosis {
		t.Fatalf("order status should transition to IN_DIAGNOSIS, got %s", updatedOrder.Status)
	}
}
