package entities

import (
	"errors"
	"strings"
)

// ProviderRequirement is the response-safe contract sent to an LLM. Provider
// URLs, credentials, and authorization headers are intentionally absent.
type ProviderRequirement struct {
	Capability        string            `json:"capability"`
	ProviderServer    string            `json:"providerServer"`
	Tool              string            `json:"tool"`
	Purpose           string            `json:"purpose"`
	Required          bool              `json:"required"`
	BeforeTransitions []string          `json:"beforeTransitions,omitempty"`
	InputMapping      map[string]string `json:"inputMapping,omitempty"`
	OutputMapping     map[string]string `json:"outputMapping,omitempty"`
	Status            string            `json:"status,omitempty"`
	Error             string            `json:"error,omitempty"`
	HardStop          bool              `json:"hardStop,omitempty"`
	NextAction        string            `json:"nextAction,omitempty"`
}

// Validate checks the metadata that is required to safely direct an LLM to a
// provider tool. It never accepts an endpoint as a substitute for an alias.
func (r ProviderRequirement) Validate() error {
	if strings.TrimSpace(r.Capability) == "" {
		return errors.New("capability is required")
	}
	if !r.Required {
		return nil
	}
	if strings.TrimSpace(r.ProviderServer) == "" {
		return errors.New("provider server alias is required")
	}
	if strings.TrimSpace(r.Tool) == "" {
		return errors.New("provider tool is required")
	}
	if strings.TrimSpace(r.Purpose) == "" {
		return errors.New("provider purpose is required")
	}
	return nil
}
