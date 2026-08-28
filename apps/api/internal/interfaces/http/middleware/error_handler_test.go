package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
)

func TestErrorHandlerRateLimited(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := domaincap.NewCapabilityError(domaincap.ErrorKindRateLimited, "capability.rate_limited", "rate limit exceeded")
	ErrorHandler(err, c)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header for 429")
	}
}

func TestErrorHandlerUnauthorizedStays403(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := domaincap.NewCapabilityError(domaincap.ErrorKindUnauthorized, "capability.unauthorized", "denied")
	ErrorHandler(err, c)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
