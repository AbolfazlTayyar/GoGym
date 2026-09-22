package httpx

// Error codes. These are the stable strings in ErrorBody.Code that the
// frontend branches on, so treat them as API surface: add new ones freely,
// but don't reword an existing one.
//
// Only the codes the current endpoints actually return are defined — a task
// that adds an endpoint adds the codes it needs, rather than this list
// growing a taxonomy for handlers nobody has written yet.
const (
	// CodeValidationFailed is a malformed or rejected request body. It is
	// the only code that populates ErrorBody.Fields.
	CodeValidationFailed = "validation_failed"
	// CodeInvalidCredentials is a failed login. Deliberately the same for
	// an unknown phone and a wrong password.
	CodeInvalidCredentials = "invalid_credentials"
	// CodeUnauthorized is a missing, malformed, or expired bearer token.
	CodeUnauthorized = "unauthorized"
	// CodeNotFound is an unknown route, or a resource the authenticated
	// coach can't see (see the 404-not-403 convention in docs/architecture.md).
	CodeNotFound = "not_found"
	// CodeConflict is a uniqueness violation, e.g. a phone already registered.
	CodeConflict = "conflict"
	// CodeRateLimited is a request rejected by the /auth/* rate limiter.
	CodeRateLimited = "rate_limited"
	// CodeMethodNotAllowed is a known path requested with the wrong method.
	// Gin answers this one without reaching a handler — see
	// internal/server/fallback.go.
	CodeMethodNotAllowed = "method_not_allowed"
	// CodeInternalError is an unexpected server-side failure. Its message is
	// deliberately generic — the detail goes to the logs, not the client.
	CodeInternalError = "internal_error"
)

// Error messages shared by more than one response.
const (
	MsgUnauthorized     = "unauthorized"
	MsgInvalidRequest   = "invalid request"
	MsgTooManyRequests  = "too many requests"
	MsgInternalError    = "internal error"
	MsgNotFound         = "not found"
	MsgMethodNotAllowed = "method not allowed"
)
