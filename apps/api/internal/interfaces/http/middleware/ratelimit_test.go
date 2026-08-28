package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// stubLimiter is a configurable RateLimiter for middleware tests.
type stubLimiter struct {
	allowed bool
	err     error
}

func (s *stubLimiter) Allow(_ context.Context, _ string) (bool, error) {
	return s.allowed, s.err
}

func newTestContext(method, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestRateLimitAllowed(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/api/auth/login", `{"email":"a@b.com"}`)
	handler := RateLimit(&stubLimiter{allowed: true}, LoginKey)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRateLimitDenied(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/api/auth/login", `{"email":"a@b.com"}`)
	handler := RateLimit(&stubLimiter{allowed: false}, LoginKey)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	err := handler(c)
	if err == nil {
		t.Fatal("expected a 429 error")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 HTTPError, got %v", err)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429")
	}
}

func TestRateLimitFailOpenOnError(t *testing.T) {
	c, _ := newTestContext(http.MethodPost, "/api/auth/login", `{"email":"a@b.com"}`)
	handler := RateLimit(&stubLimiter{err: errors.New("limiter down")}, LoginKey)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	// Fail-open: request proceeds despite limiter error.
	if err := handler(c); err != nil {
		t.Fatalf("expected fail-open (no error), got %v", err)
	}
}

func TestRegisterKeyUsesIP(t *testing.T) {
	c, _ := newTestContext(http.MethodPost, "/api/auth/register", `{"email":"x@y.com"}`)
	c.Request().RemoteAddr = "203.0.113.9:1234"
	key, ok := RegisterKey(c)
	if !ok || !strings.HasPrefix(key, "route:register:ip:") {
		t.Fatalf("expected register IP key, got %q ok=%v", key, ok)
	}
	if !strings.HasSuffix(key, "203.0.113.9") {
		t.Fatalf("expected client IP in key, got %q", key)
	}
}

func TestLoginKeyUsesEmailAndRestoresBody(t *testing.T) {
	c, _ := newTestContext(http.MethodPost, "/api/auth/login", `{"email":"user@example.com"}`)
	key, ok := LoginKey(c)
	if !ok || key != "route:login:email:user@example.com" {
		t.Fatalf("expected email key, got %q ok=%v", key, ok)
	}
	// Body must be restorable for the controller to bind it later.
	buf := make([]byte, len(`{"email":"user@example.com"}`))
	if n, _ := c.Request().Body.Read(buf); n == 0 {
		t.Error("expected request body to remain readable after key extraction")
	}
}
