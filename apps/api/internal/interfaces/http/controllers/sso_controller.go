package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

const (
	ssoStateCookie    = "openstate_sso_state"
	ssoVerifierCookie = "openstate_sso_verifier"
	ssoCookieMaxAge   = 10 * 60 // 10 minutes
)

// SSOController exposes the OIDC SSO login flow (PRD §79): initiate the
// authorization URL (persisting state + PKCE verifier in cookies) and handle the
// provider callback. It contains no business logic.
type SSOController struct {
	svc *appservices.SSOService
	// frontendBaseURL is where the callback redirects after login.
	frontendBaseURL string
}

// NewSSOController builds an SSOController.
func NewSSOController(svc *appservices.SSOService, frontendBaseURL string) *SSOController {
	return &SSOController{svc: svc, frontendBaseURL: frontendBaseURL}
}

// Start redirects the user to the provider authorization URL.
func (ctrl *SSOController) Start(c echo.Context) error {
	provider := c.Param("provider")
	authURL, state, verifier, err := ctrl.svc.StartAuth(provider)
	if err != nil {
		return err
	}

	setSSOCookie(c, ssoStateCookie, state)
	setSSOCookie(c, ssoVerifierCookie, verifier)

	return c.Redirect(http.StatusFound, authURL)
}

// Callback handles the provider redirect: verifies state, exchanges the code,
// and redirects the frontend with the access token (or an error).
func (ctrl *SSOController) Callback(c echo.Context) error {
	provider := c.Param("provider")
	state := c.QueryParam("state")
	code := c.QueryParam("code")

	// State (CSRF) verification.
	savedState := readSSOCookie(c, ssoStateCookie)
	if state == "" || savedState != state {
		clearSSOCookies(c)
		return c.Redirect(http.StatusFound, ctrl.frontendBaseURL+"?sso=error&reason=invalid_state")
	}

	verifier := readSSOCookie(c, ssoVerifierCookie)
	clearSSOCookies(c)

	result, err := ctrl.svc.CompleteLogin(c.Request().Context(), provider, code, verifier)
	if err != nil {
		var de *domain.DomainError
		if asDomain(err, &de) {
			return c.Redirect(http.StatusFound, ctrl.frontendBaseURL+"?sso=error&reason="+de.Code)
		}
		return c.Redirect(http.StatusFound, ctrl.frontendBaseURL+"?sso=error&reason=internal")
	}

	return c.Redirect(http.StatusFound, ctrl.frontendBaseURL+"?sso=success&token="+result.AccessToken)
}

// Providers lists the enabled SSO providers for the frontend.
func (ctrl *SSOController) Providers(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"data": ctrl.svc.EnabledProviders()})
}

func setSSOCookie(c echo.Context, name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   ssoCookieMaxAge,
	}
	c.SetCookie(cookie)
}

func readSSOCookie(c echo.Context, name string) string {
	if cookie, err := c.Cookie(name); err == nil {
		return cookie.Value
	}
	return ""
}

func clearSSOCookies(c echo.Context) {
	for _, name := range []string{ssoStateCookie, ssoVerifierCookie} {
		c.SetCookie(&http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0)})
	}
}

func asDomain(err error, target **domain.DomainError) bool {
	de, ok := err.(*domain.DomainError)
	if ok {
		*target = de
	}
	return ok
}
