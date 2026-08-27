package capability

// ErrorKind classifies a capability failure (PRD §87).
type ErrorKind string

const (
	ErrorKindTimeout        ErrorKind = "TIMEOUT"
	ErrorKindUnauthorized   ErrorKind = "UNAUTHORIZED"
	ErrorKindValidation     ErrorKind = "VALIDATION"
	ErrorKindUnavailable    ErrorKind = "UNAVAILABLE"
	ErrorKindBusiness       ErrorKind = "BUSINESS"
	ErrorKindExternal       ErrorKind = "EXTERNAL"
	ErrorKindInternal       ErrorKind = "INTERNAL"
)

// CapabilityError is a classified capability failure carrying a kind (PRD §87)
// and a deterministic event code (PRD §63). Raw provider errors are never
// exposed to callers (PRD §2951).
type CapabilityError struct {
	Kind    ErrorKind
	Code    string
	Message string
}

func (e *CapabilityError) Error() string { return e.Message }

// NewCapabilityError builds a classified capability error.
func NewCapabilityError(kind ErrorKind, code, message string) *CapabilityError {
	return &CapabilityError{Kind: kind, Code: code, Message: message}
}

// CodeForCapabilityEvent maps an ErrorKind to its deterministic capability
// event code (PRD §63).
func CodeForCapabilityEvent(kind ErrorKind) string {
	switch kind {
	case ErrorKindTimeout:
		return "capability.timeout"
	case ErrorKindUnauthorized:
		return "capability.unauthorized"
	case ErrorKindValidation:
		return "capability.validation_failed"
	case ErrorKindUnavailable:
		return "capability.unavailable"
	case ErrorKindBusiness:
		return "capability.business_error"
	default:
		return "capability.failed"
	}
}

// Retryable reports whether an error kind is eligible for retry (PRD §88).
func (e *CapabilityError) Retryable() bool {
	return e.Kind == ErrorKindTimeout || e.Kind == ErrorKindUnavailable || e.Kind == ErrorKindExternal
}
