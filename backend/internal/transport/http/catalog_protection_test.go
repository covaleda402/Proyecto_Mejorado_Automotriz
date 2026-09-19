package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type fakeCustomerStore struct{}

func (f *fakeCustomerStore) Save(_ context.Context, _ domain.Customer) error { return nil }
func (f *fakeCustomerStore) FindByID(_ context.Context, _ string) (domain.Customer, error) {
	return domain.Customer{}, domain.ErrNotFound
}
func (f *fakeCustomerStore) FindByDocumentNumber(_ context.Context, _ string) (domain.Customer, error) {
	return domain.Customer{}, domain.ErrNotFound
}
func (f *fakeCustomerStore) List(_ context.Context) ([]domain.Customer, error) {
	return []domain.Customer{}, nil
}

func TestCatalogListEndpointsProtectedByRequireAdministrator(t *testing.T) {
	// CustomerHandler.List
	custHandler := NewCustomerHandler(usecase.NewCustomerUseCase(&fakeCustomerStore{}, sequentialIDTest(), testClock()))
	// VehicleHandler.List
	vehHandler := NewVehicleHandler(usecase.NewVehicleUseCase(newFakeVehicleStore("v1"), &fakeCustomerStore{}, sequentialIDTest(), testClock()))
	// WarrantyHandler.List
	warHandler := newWarrantyHandler(t, newFakeWarrantyStore())
	// TechnicianHandler.List
	techHandler := NewTechnicianHandler(usecase.NewTechnicianUseCase(&fakeTechnicianStore{}, sequentialIDTest(), testClock()))

	testCases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		path    string
	}{
		{"CustomerList", custHandler.List, "/api/customer"},
		{"VehicleList", vehHandler.List, "/api/vehicle"},
		{"WarrantyList", warHandler.List, "/api/warranty"},
		{"TechnicianList", techHandler.List, "/api/technician"},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"_ForbiddenForTechnician", func(t *testing.T) {
			req := asCaller(httptest.NewRequest(http.MethodGet, tc.path, nil), domain.RoleTechnician)
			rec := httptest.NewRecorder()
			tc.handler(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected 403 Forbidden for %s with technician role, got %d", tc.name, rec.Code)
			}
		})

		t.Run(tc.name+"_AllowedForAdministrator", func(t *testing.T) {
			req := asCaller(httptest.NewRequest(http.MethodGet, tc.path, nil), domain.RoleAdministrator)
			rec := httptest.NewRecorder()
			tc.handler(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 OK for %s with administrator role, got %d", tc.name, rec.Code)
			}
		})
	}
}

func sequentialIDTest() func() string {
	return func() string { return "id-test" }
}

func (f *fakeTechnicianStore) ReleaseByServiceOrder(_ context.Context, _ string, _ time.Time) error {
	return nil
}
