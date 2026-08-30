package controllers

import (
	"net/http"
	"strconv"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

// ProjectHeader carries the project id for project-scoped workflow operations.
// It is optional: when absent, the service falls back to the tenant's default
// project (PRD §3.1.1). The tenant always comes from X-Tenant-ID.
const ProjectHeader = "X-Project-ID"

// WorkflowController exposes the Builder API (PRD 146): workflow definition draft
// CRUD, publish, and version listing. It parses requests, calls the service, and
// formats responses. It contains no business logic (PRD §74).
type WorkflowController struct {
	svc        *appservices.BuilderService
	simulation *appservices.SimulationService
}

// NewWorkflowController builds a WorkflowController.
func NewWorkflowController(svc *appservices.BuilderService, simulation ...*appservices.SimulationService) *WorkflowController {
	sim := appservices.NewSimulationService()
	if len(simulation) > 0 && simulation[0] != nil {
		sim = simulation[0]
	}
	return &WorkflowController{svc: svc, simulation: sim}
}

func (ctrl *WorkflowController) Create(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.CreateWorkflowRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.svc.CreateDraft(c.Request().Context(), tenantID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": result})
}

// Simulate runs an unsaved workflow snapshot in the transient sandbox.
func (ctrl *WorkflowController) Simulate(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.SimulateWorkflowRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.simulation.Simulate(c.Request().Context(), tenantID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func (ctrl *WorkflowController) List(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.List(c.Request().Context(), tenantID, c.QueryParam("projectId"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result.Data})
}

func (ctrl *WorkflowController) Get(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.Get(c.Request().Context(), tenantID, c.Request().Header.Get(ProjectHeader), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func (ctrl *WorkflowController) Update(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.UpdateWorkflowRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.svc.UpdateDraft(c.Request().Context(), tenantID, c.Request().Header.Get(ProjectHeader), c.Param("id"), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func (ctrl *WorkflowController) Publish(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	var req dtos.PublishWorkflowRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.svc.Publish(c.Request().Context(), tenantID, c.Request().Header.Get(ProjectHeader), c.Param("id"), actor, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": result})
}

func (ctrl *WorkflowController) ListVersions(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.ListVersions(c.Request().Context(), tenantID, c.Request().Header.Get(ProjectHeader), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

// GetVersion returns one immutable workflow version with its full definition.
func (ctrl *WorkflowController) GetVersion(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	versionNo, err := strconv.Atoi(c.Param("versionNo"))
	if err != nil || versionNo < 1 {
		return domain.NewValidation("versionNo must be a positive integer")
	}
	result, err := ctrl.svc.GetVersion(c.Request().Context(), tenantID, c.Request().Header.Get(ProjectHeader), c.Param("id"), versionNo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

// CompareVersions returns a deterministic semantic diff between two immutable
// versions of one workflow.
func (ctrl *WorkflowController) CompareVersions(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	baseVersion, err := strconv.Atoi(c.QueryParam("baseVersion"))
	if err != nil {
		return domain.NewValidation("baseVersion must be a positive integer")
	}
	targetVersion, err := strconv.Atoi(c.QueryParam("targetVersion"))
	if err != nil {
		return domain.NewValidation("targetVersion must be a positive integer")
	}
	result, err := ctrl.svc.CompareVersions(c.Request().Context(), tenantID, c.Request().Header.Get(ProjectHeader), c.Param("id"), baseVersion, targetVersion)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}
