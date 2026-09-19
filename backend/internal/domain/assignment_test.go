package domain_test

import (
	"errors"
	"testing"
	"time"

	"workshop/internal/domain"
)

func TestNewAssignmentSetsActiveMarkers(t *testing.T) {
	now := time.Now()
	a, err := domain.NewAssignment("assign-1", "order-1", "tech-1", now)
	if err != nil {
		t.Fatalf("unexpected error creating assignment: %v", err)
	}

	if !a.IsActive {
		t.Error("expected assignment to be active")
	}
	if a.ActiveMarker == nil || *a.ActiveMarker != "tech-1" {
		t.Errorf("expected ActiveMarker to be 'tech-1', got %v", a.ActiveMarker)
	}
	if a.ActiveOrderMarker == nil || *a.ActiveOrderMarker != "order-1" {
		t.Errorf("expected ActiveOrderMarker to be 'order-1', got %v", a.ActiveOrderMarker)
	}
	if a.ReleasedAt != nil {
		t.Errorf("expected ReleasedAt to be nil, got %v", a.ReleasedAt)
	}

	// Test Release
	releaseTime := now.Add(time.Hour)
	a.Release(releaseTime)

	if a.IsActive {
		t.Error("expected assignment to be inactive after release")
	}
	if a.ActiveMarker != nil {
		t.Errorf("expected ActiveMarker to be nil after release, got %v", a.ActiveMarker)
	}
	if a.ActiveOrderMarker != nil {
		t.Errorf("expected ActiveOrderMarker to be nil after release, got %v", a.ActiveOrderMarker)
	}
	if a.ReleasedAt == nil || *a.ReleasedAt != releaseTime {
		t.Errorf("expected ReleasedAt to be %v, got %v", releaseTime, a.ReleasedAt)
	}
}

func TestNewAssignmentValidations(t *testing.T) {
	now := time.Now()
	if _, err := domain.NewAssignment("", "order-1", "tech-1", now); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty id, got %v", err)
	}
	if _, err := domain.NewAssignment("assign-1", "", "tech-1", now); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty order id, got %v", err)
	}
	if _, err := domain.NewAssignment("assign-1", "order-1", "", now); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty technician id, got %v", err)
	}
}
