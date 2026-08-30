package controllers

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/irosadie/open-state/api/internal/application/services"
	domainengine "github.com/irosadie/open-state/api/internal/domain/engine"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

func TestWorkflowSimulationRejectsMissingTenant(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/api/workflows/simulate", strings.NewReader(`{"definition":{}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx := e.NewContext(req, httptest.NewRecorder())

	err := NewWorkflowController(nil, services.NewSimulationService()).Simulate(ctx)
	if err == nil {
		t.Fatal("expected missing tenant error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrUnauthorized {
		t.Fatalf("expected unauthorized domain error, got %v", err)
	}
}

func TestWorkflowSimulationReturnsDataEnvelope(t *testing.T) {
	definition, err := json.Marshal(domainengine.WorkflowDefinition{
		Slug:        "simulation",
		EntryNodeID: "start",
		Nodes:       []domainengine.WorkflowNode{{ID: "start", Kind: domainengine.NodeKindStart, Name: "START"}},
	})
	if err != nil {
		t.Fatalf("marshal definition: %v", err)
	}
	body, _ := json.Marshal(map[string]json.RawMessage{"definition": definition})
	e := echo.New()
	req := httptest.NewRequest("POST", "/api/workflows/simulate", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	recorder := httptest.NewRecorder()
	ctx := e.NewContext(req, recorder)

	if err := NewWorkflowController(nil, services.NewSimulationService()).Simulate(ctx); err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if recorder.Code != 200 || !strings.Contains(recorder.Body.String(), `"data"`) {
		t.Fatalf("expected 200 data envelope, got %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestWorkflowSimulationRejectsMalformedPayload(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/api/workflows/simulate", strings.NewReader(`{"definition":`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	ctx := e.NewContext(req, httptest.NewRecorder())

	err := NewWorkflowController(nil, services.NewSimulationService()).Simulate(ctx)
	if err == nil {
		t.Fatal("expected malformed payload error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrValidation {
		t.Fatalf("expected validation domain error, got %v", err)
	}
}

func TestWorkflowVersionCompareRejectsMalformedInputs(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/api/workflows/wf-1/versions/compare?baseVersion=one&targetVersion=2", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	ctx := e.NewContext(req, httptest.NewRecorder())

	err := NewWorkflowController(nil).CompareVersions(ctx)
	if err == nil {
		t.Fatal("expected malformed baseVersion error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestWorkflowVersionDetailRejectsNonPositiveVersion(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/api/workflows/wf-1/versions/0", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	ctx := e.NewContext(req, httptest.NewRecorder())
	ctx.SetPath("/api/workflows/:id/versions/:versionNo")
	ctx.SetParamNames("id", "versionNo")
	ctx.SetParamValues("wf-1", "0")

	err := NewWorkflowController(nil).GetVersion(ctx)
	if err == nil {
		t.Fatal("expected non-positive version error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}
