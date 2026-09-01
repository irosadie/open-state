package controllers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

func TestIntentControllerListRequiresTenant(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/api/intents", nil)
	recorder := httptest.NewRecorder()
	ctx := e.NewContext(req, recorder)

	err := NewIntentController(nil).List(ctx)
	if err == nil || !strings.Contains(err.Error(), "missing tenant header") {
		t.Fatalf("expected missing tenant header error, got %v", err)
	}
}

func TestIntentControllerListReturnsDataEnvelope(t *testing.T) {
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const projectID = "00000000-0000-0000-0000-000000000006"
	projects := &controllerProjectRepository{projectID: projectID, tenantID: tenantID}
	intents := &controllerIntentRepository{}
	intentService := services.NewIntentService(intents, nil)
	catalogService := services.NewIntentCatalogService(intentService, projects)

	e := echo.New()
	req := httptest.NewRequest("GET", "/api/intents", nil)
	req.Header.Set("X-Tenant-ID", tenantID)
	recorder := httptest.NewRecorder()
	ctx := e.NewContext(req, recorder)

	if err := NewIntentController(catalogService).List(ctx); err != nil {
		t.Fatalf("list intents: %v", err)
	}
	if recorder.Code != 200 || !strings.Contains(recorder.Body.String(), `"data"`) {
		t.Fatalf("expected 200 data envelope, got %d %s", recorder.Code, recorder.Body.String())
	}
}

// These minimal ports keep the controller test focused on HTTP parsing and
// response formatting; service behavior is covered in application tests.
type controllerProjectRepository struct {
	projectID string
	tenantID  string
}

func (r *controllerProjectRepository) Create(_ context.Context, _, _, _ string, _ entities.ProjectStatus) (*entities.Project, error) {
	return nil, nil
}

func (r *controllerProjectRepository) FindByID(_ context.Context, tenantID, projectID string) (*entities.Project, error) {
	if tenantID != r.tenantID || projectID != r.projectID {
		return nil, domain.NewNotFound("project not found")
	}
	return &entities.Project{ID: r.projectID, TenantID: r.tenantID}, nil
}

func (r *controllerProjectRepository) FindBySlug(_ context.Context, tenantID, slug string) (*entities.Project, error) {
	if tenantID != r.tenantID || slug != "default" {
		return nil, domain.NewNotFound("project not found")
	}
	return &entities.Project{ID: r.projectID, TenantID: r.tenantID}, nil
}

func (r *controllerProjectRepository) ListByTenant(context.Context, string) ([]entities.Project, error) {
	return nil, nil
}

type controllerIntentRepository struct{}

func (*controllerIntentRepository) ListRoutable(context.Context, string, string) ([]entities.Intent, error) {
	return []entities.Intent{{Key: "BOOKING_PADEL", ProjectID: "00000000-0000-0000-0000-000000000006"}}, nil
}

func (*controllerIntentRepository) FindRoutable(context.Context, string, string, string) (*entities.Intent, error) {
	return nil, domain.NewNotFound("intent not found")
}

func (*controllerIntentRepository) Upsert(context.Context, string, string, string, string, string, string, []string) (*entities.Intent, error) {
	return nil, nil
}
