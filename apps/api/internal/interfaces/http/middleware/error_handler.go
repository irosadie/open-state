package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	domain "github.com/vibecoding-starter/go-shared/domain"
)

// errorCodeToHTTP maps DomainError codes to HTTP status codes.
var errorCodeToHTTP = map[string]int{
	domain.ErrNotFound:     http.StatusNotFound,
	domain.ErrUnauthorized: http.StatusUnauthorized,
	domain.ErrForbidden:    http.StatusForbidden,
	domain.ErrConflict:     http.StatusConflict,
	domain.ErrValidation:   http.StatusUnprocessableEntity,
	domain.ErrInternal:     http.StatusInternalServerError,
}

// ErrorHandler is the centralized Echo error handler.
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var status int
	var message string

	if de, ok := err.(*domain.DomainError); ok {
		status = errorCodeToHTTP[de.Code]
		if status == 0 {
			status = http.StatusInternalServerError
		}
		message = de.Message
	} else if he, ok := err.(*echo.HTTPError); ok {
		status = he.Code
		message = http.StatusText(he.Code)
	} else {
		status = http.StatusInternalServerError
		message = "internal server error"
	}

	_ = c.JSON(status, map[string]string{"error": message})
}
