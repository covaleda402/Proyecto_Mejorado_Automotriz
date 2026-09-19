package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

func newTestTechnicianHandler(t *testing.T, store *fakeTechnicianStore) TechnicianHandler {
	t.Helper()
	return NewTechnicianHandler(usecase.NewTechnicianUseCase(store, sequentialIDTest(), testClock()))
}

func TestTechnicianHandler_Create_Success(t *testing.T) {
	store := newFakeTechnicianStore(t, "tech-1", "user-tech-1")
	handler := newTestTechnicianHandler(t, store)

	body, _ := json.Marshal(createTechnicianRequest{
		FullName:  "Carlos Mendoza",
		Username:  "cmendoza",
		Password:  "Password#2026!",
		Specialty: "Transmisiones y cajas",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/technician", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), callerContextKey, caller{
		UserID: "admin-1",
		Role:   domain.RoleAdministrator,
	}))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp technicianResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.FullName != "Carlos Mendoza" {
		t.Errorf("expected FullName 'Carlos Mendoza', got %q", resp.FullName)
	}
	if !resp.IsActive {
		t.Error("newly created technician must be active")
	}
	if !resp.CanReceiveAssignment {
		t.Error("newly created technician must be eligible to receive assignment")
	}
}

func TestTechnicianHandler_Create_ForbiddenForTechnician(t *testing.T) {
	store := newFakeTechnicianStore(t, "tech-1", "user-tech-1")
	handler := newTestTechnicianHandler(t, store)

	body, _ := json.Marshal(createTechnicianRequest{
		FullName:  "Otro Tecnico",
		Username:  "otecnico",
		Password:  "Password#2026!",
		Specialty: "Frenos",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/technician", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), callerContextKey, caller{
		UserID: "user-tech-1",
		Role:   domain.RoleTechnician,
	}))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for technician caller, got %d", rec.Code)
	}
}

func TestTechnicianHandler_SetAccess_SuccessAndIdempotent(t *testing.T) {
	store := newFakeTechnicianStore(t, "tech-1", "user-tech-1")
	handler := newTestTechnicianHandler(t, store)

	body, _ := json.Marshal(setAccessRequest{Active: false})
	req := httptest.NewRequest(http.MethodPatch, "/api/technician/tech-1/access", bytes.NewReader(body))
	req.SetPathValue("technicianId", "tech-1")
	req = req.WithContext(context.WithValue(req.Context(), callerContextKey, caller{
		UserID: "admin-1",
		Role:   domain.RoleAdministrator,
	}))
	rec := httptest.NewRecorder()

	handler.SetAccess(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp technicianResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.IsActive {
		t.Error("technician must be inactive after deactivation")
	}
	if resp.CanReceiveAssignment {
		t.Error("inactive technician must not be assignable")
	}

	// Idempotency: call again with active: false
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPatch, "/api/technician/tech-1/access", bytes.NewReader(body))
	req2.SetPathValue("technicianId", "tech-1")
	req2 = req2.WithContext(req.Context())
	handler.SetAccess(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on idempotent retry, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

func TestTechnicianHandler_SetAccess_SelfDeactivationForbidden(t *testing.T) {
	// The acting administrator cannot deactivate themselves (BR-ACTIVE-07)
	store := newFakeTechnicianStore(t, "admin-tech", "admin-1")
	handler := newTestTechnicianHandler(t, store)

	body, _ := json.Marshal(setAccessRequest{Active: false})
	req := httptest.NewRequest(http.MethodPatch, "/api/technician/admin-tech/access", bytes.NewReader(body))
	req.SetPathValue("technicianId", "admin-tech")
	req = req.WithContext(context.WithValue(req.Context(), callerContextKey, caller{
		UserID: "admin-1",
		Role:   domain.RoleAdministrator,
	}))
	rec := httptest.NewRecorder()

	handler.SetAccess(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden on self-deactivation, got %d: %s", rec.Code, rec.Body.String())
	}
}
