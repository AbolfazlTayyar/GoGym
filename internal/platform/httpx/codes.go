package httpx

// Error codes are API surface the frontend branches on: add new ones, never reword existing ones.
const (
	// CodeValidationFailed is the only code that populates ErrorBody.Fields.
	CodeValidationFailed = "validation_failed"
	// CodeInvalidCredentials is deliberately the same for an unknown phone and a wrong password.
	CodeInvalidCredentials = "invalid_credentials"
	CodeUnauthorized       = "unauthorized"
	// CodeNotFound also covers resources the coach can't see (404, not 403).
	CodeNotFound = "not_found"
	// CodeForbidden is for a resource the coach can see but not change, such as a universal movement.
	CodeForbidden        = "forbidden"
	CodeConflict         = "conflict"
	CodeRateLimited      = "rate_limited"
	CodeMethodNotAllowed = "method_not_allowed"
	CodeInternalError    = "internal_error"
)

const (
	MsgUnauthorized    = "unauthorized"
	MsgInvalidRequest  = "invalid request"
	MsgTooManyRequests = "too many requests"
	MsgInternalError   = "internal error"
	// MsgNotFound also answers an id in a request body the coach can't use, without saying whether it exists.
	MsgNotFound           = "not found"
	MsgMethodNotAllowed   = "method not allowed"
	MsgInvalidCredentials = "invalid credentials"
)

// Field messages, here and in each module, are matched word for word by web/src/api/fields.ts: reword one only together with its rule there.
const (
	// MsgFieldRequired is shared by binding-tag and service-level validation so a missing field reads the same either way.
	MsgFieldRequired = "is required"
	MsgDateFormat    = "must be a date in YYYY-MM-DD format"
	// MsgMaxLength takes the limit as an int from services or as a string from binding tags, hence %v.
	MsgMaxLength = "must be at most %v characters"
)
