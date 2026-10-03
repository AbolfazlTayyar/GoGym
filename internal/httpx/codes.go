package httpx

// Error codes are API surface the frontend branches on: add new ones, never reword existing ones.
const (
	// CodeValidationFailed is the only code that populates ErrorBody.Fields.
	CodeValidationFailed = "validation_failed"
	// CodeInvalidCredentials is deliberately the same for an unknown phone and a wrong password.
	CodeInvalidCredentials = "invalid_credentials"
	CodeUnauthorized       = "unauthorized"
	// CodeNotFound also covers resources the coach can't see (404, not 403).
	CodeNotFound         = "not_found"
	CodeConflict         = "conflict"
	CodeRateLimited      = "rate_limited"
	CodeMethodNotAllowed = "method_not_allowed"
	CodeInternalError    = "internal_error"
)

const (
	MsgUnauthorized     = "unauthorized"
	MsgInvalidRequest   = "invalid request"
	MsgTooManyRequests  = "too many requests"
	MsgInternalError    = "internal error"
	MsgNotFound         = "not found"
	MsgMethodNotAllowed = "method not allowed"
)

// MsgFieldRequired is shared by binding-tag and service-level validation so a missing field reads the same either way.
const MsgFieldRequired = "is required"
