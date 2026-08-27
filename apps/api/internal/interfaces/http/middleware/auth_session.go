package middleware

import (
	"github.com/labstack/echo/v4"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// AuthSession validates that the JWT token has an active DB session.
func AuthSession(repo repositories.IAuthRepository, tokenSvc interface{ HashToken(string) string }) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("token").(string)
			if !ok || token == "" {
				return domain.NewUnauthorized("missing token")
			}

			tokenHash := tokenSvc.HashToken(token)
			session, err := repo.FindSessionByTokenHash(c.Request().Context(), tokenHash)
			if err != nil || session == nil {
				return domain.NewUnauthorized("session expired or invalid")
			}

			return next(c)
		}
	}
}
