package controllers

import (
	"net/http"
	"time"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/labstack/echo/v4"
)

// AuditController exposes the tenant-scoped audit trail query API (PRD 50),
// protected by the audit:read permission guard. It parses request filters and
// delegates to the audit service; it contains no business logic.
type AuditController struct {
	svc *appservices.AuditService
}

// NewAuditController builds an AuditController.
func NewAuditController(svc *appservices.AuditService) *AuditController {
	return &AuditController{svc: svc}
}

// List returns a filtered, paginated view of the tenant's audit trail.
func (ctrl *AuditController) List(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}

	q := appservices.AuditQuery{
		Page:     parseInt(c.QueryParam("page"), 1),
		PageSize: parseInt(c.QueryParam("pageSize"), 0),
	}

	if v := c.QueryParam("action"); v != "" {
		a := entities.AuditAction(v)
		q.Action = &a
	}
	if v := c.QueryParam("resourceType"); v != "" {
		q.ResourceType = &v
	}
	if v := c.QueryParam("resourceId"); v != "" {
		q.ResourceID = &v
	}
	if v := c.QueryParam("actor"); v != "" {
		q.Actor = &v
	}
	if v := c.QueryParam("correlationId"); v != "" {
		q.CorrelationID = &v
	}
	if v := c.QueryParam("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.From = &t
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.To = &t
		}
	}

	result, err := ctrl.svc.List(c.Request().Context(), tenantID, q)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

// parseInt parses a string into an int with a fallback default on error/empty.
func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	return n
}
