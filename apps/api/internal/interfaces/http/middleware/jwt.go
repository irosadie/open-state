package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	domain "github.com/vibecoding-starter/go-shared/domain"
	"github.com/vibecoding-starter/api/internal/domain/services"
)

const UserIDKey = "userID"

// JWT validates the Bearer token and sets userID in context.
func JWT(tokenSvc services.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return domain.NewUnauthorized("missing or invalid authorization header")
			}

			token := strings.TrimPrefix(header, "Bearer ")
			userID, err := tokenSvc.ValidateToken(token)
			if err != nil {
				return domain.NewUnauthorized("invalid or expired token")
			}

			c.Set(UserIDKey, userID)
			c.Set("token", token)
			return next(c)
		}
	}
}

// RequireAuth is a shorthand guard that returns 401 if userID not set.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(UserIDKey) == nil {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "unauthorized"}
		}
		return next(c)
	}
}
