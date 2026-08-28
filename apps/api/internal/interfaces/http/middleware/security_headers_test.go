package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSecurityHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := SecurityHeaders(SecurityHeadersConfig{EnableHSTS: true, HSTSMaxAge: 1000})(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if err := h(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	for _, header := range []string{"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Permissions-Policy", "Content-Security-Policy", "Strict-Transport-Security"} {
		if rec.Header().Get(header) == "" {
			t.Errorf("expected header %s to be set", header)
		}
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("nosniff = %q", rec.Header().Get("X-Content-Type-Options"))
	}
	if rec.Header().Get("Strict-Transport-Security") != "max-age=1000; includeSubDomains" {
		t.Errorf("HSTS = %q", rec.Header().Get("Strict-Transport-Security"))
	}
}

func TestSecurityHeadersHSTSOff(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := SecurityHeaders(SecurityHeadersConfig{EnableHSTS: false})(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	_ = h(c)

	if rec.Header().Get("Strict-Transport-Security") != "" {
		t.Error("expected no HSTS when disabled")
	}
}
