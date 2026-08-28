package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
	domain "github.com/irosadie/open-state/go-shared/domain"
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

// capabilityKindToHTTP maps a classified capability error kind to an HTTP status
// code (PRD §87, §2064).
var capabilityKindToHTTP = map[domaincap.ErrorKind]int{
	domaincap.ErrorKindValidation:   http.StatusUnprocessableEntity,
	domaincap.ErrorKindUnauthorized: http.StatusForbidden,
	domaincap.ErrorKindTimeout:      http.StatusInternalServerError,
	domaincap.ErrorKindUnavailable:  http.StatusInternalServerError,
	domaincap.ErrorKindExternal:     http.StatusInternalServerError,
	domaincap.ErrorKindBusiness:     http.StatusUnprocessableEntity,
	domaincap.ErrorKindInternal:     http.StatusInternalServerError,
}

// classifiedCapabilityError is the serialized shape of a classified capability
// failure returned to callers (kind/code/message, never a raw provider error).
type classifiedCapabilityError struct {
	Kind    domaincap.ErrorKind `json:"kind"`
	Code    string              `json:"code"`
	Message string              `json:"message"`
}

// ErrorHandler is the centralized Echo error handler.
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var status int
	var body any

	if ce, ok := err.(*domaincap.CapabilityError); ok {
		status = capabilityKindToHTTP[ce.Kind]
		if status == 0 {
			status = http.StatusInternalServerError
		}
		body = map[string]classifiedCapabilityError{
			"error": {
				Kind:    ce.Kind,
				Code:    ce.Code,
				Message: ce.Message,
			},
		}
	} else if de, ok := err.(*domain.DomainError); ok {
		status = errorCodeToHTTP[de.Code]
		if status == 0 {
			status = http.StatusInternalServerError
		}
		body = map[string]string{"error": de.Message}
	} else if he, ok := err.(*echo.HTTPError); ok {
		status = he.Code
		body = map[string]string{"error": http.StatusText(he.Code)}
	} else {
		status = http.StatusInternalServerError
		body = map[string]string{"error": "internal server error"}
	}

	_ = c.JSON(status, body)
}
