package usecase_test

import (
	"context"
	"errors"
	"testing"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

func TestTechnicianUseCase_Create_Success(t *testing.T) {
	repo := newFakeTechnicianRepository()
	uc := usecase.NewTechnicianUseCase(repo, sequentialID(), fixedClock())

	created, err := uc.Create(
		context.Background(),
		"Carlos Mendoza",
		"cmendoza",
		"Password#2026!",
		"Transmisiones automáticas",
	)
	if err != nil {
		t.Fatalf("expected successful creation, got: %v", err)
	}

	if created.Specialty != "Transmisiones automáticas" {
		t.Errorf("expected specialty 'Transmisiones automáticas', got %q", created.Specialty)
	}

	stored, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected technician to be persisted in repo: %v", err)
	}
	if stored.ID != created.ID {
		t.Errorf("stored ID %q does not match created ID %q", stored.ID, created.ID)
	}
}

func TestTechnicianUseCase_Create_ValidationFailures(t *testing.T) {
	repo := newFakeTechnicianRepository()
	uc := usecase.NewTechnicianUseCase(repo, sequentialID(), fixedClock())

	testCases := []struct {
		name      string
		fullName  string
		username  string
		password  string
		specialty string
	}{
		{"ShortFullName", "ab", "cmendoza", "Password#2026!", "Frenos"},
		{"ShortUsername", "Carlos Mendoza", "ab", "Password#2026!", "Frenos"},
		{"ShortPassword", "Carlos Mendoza", "cmendoza", "1234567", "Frenos"},
		{"ShortSpecialty", "Carlos Mendoza", "cmendoza", "Password#2026!", "ab"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Create(context.Background(), tc.fullName, tc.username, tc.password, tc.specialty)
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got: %v", err)
			}
		})
	}
}

func TestTechnicianUseCase_SetAccess_SuccessAndInvariants(t *testing.T) {
	tech := buildTechnician(t, "tech-1", "user-tech-1")
	repo := newFakeTechnicianRepository(tech)
	uc := usecase.NewTechnicianUseCase(repo, sequentialID(), fixedClock())

	// Deactivate
	workload, err := uc.SetAccess(context.Background(), "admin-user", "tech-1", false)
	if err != nil {
		t.Fatalf("expected successful deactivation, got: %v", err)
	}
	if workload.IsActive {
		t.Error("workload should report IsActive = false")
	}

	// Self-deactivation forbidden
	_, err = uc.SetAccess(context.Background(), "user-tech-1", "tech-1", false)
	if !errors.Is(err, domain.ErrSelfDeactivation) {
		t.Errorf("expected ErrSelfDeactivation, got: %v", err)
	}
}
