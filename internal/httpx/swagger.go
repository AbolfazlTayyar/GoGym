package httpx

// Swagger-only types: one Envelope schema can't say "error is null on success", so each state gets its own.
// swag can't emit a null example, so `make swagger` rewrites the x-null-example marker into one;
// without it Swagger UI renders these untyped fields as "string". SuccessEnvelope.Data must not get one:
// Swagger UI merges it into each route's composed data and shows null instead of the payload.

// SuccessEnvelope takes the payload via swag composition: httpx.SuccessEnvelope{data=...}.
// Meta is left out so it only appears on list routes that compose it in: httpx.SuccessEnvelope{data=...,meta=...}.
type SuccessEnvelope struct {
	Success bool `json:"success" example:"true"`
	Data    any  `json:"data"`
	Error   any  `json:"error" extensions:"x-null-example"`
}

// NoContentEnvelope documents httpx.NoContent.
type NoContentEnvelope struct {
	Success bool `json:"success" example:"true"`
	Data    any  `json:"data" extensions:"x-null-example"`
	Error   any  `json:"error" extensions:"x-null-example"`
}

// Error envelopes come one per code because swag can't attach an example to a single response;
// a shared ErrorBody showed validation_failed and a fields map on every status.

type errorEnvelope struct {
	Success bool `json:"success" example:"false"`
	Data    any  `json:"data" extensions:"x-null-example"`
}

// ValidationErrorEnvelope's fields are absent when the body isn't valid JSON or a path id is malformed.
type ValidationErrorEnvelope struct {
	errorEnvelope
	Error struct {
		Code    string            `json:"code" example:"validation_failed"`
		Message string            `json:"message" example:"invalid request"`
		Fields  map[string]string `json:"fields,omitempty" example:"<field>:is required"`
	} `json:"error"`
}

type InvalidCredentialsEnvelope struct {
	errorEnvelope
	Error struct {
		Code    string `json:"code" example:"invalid_credentials"`
		Message string `json:"message" example:"invalid credentials"`
	} `json:"error"`
}

type UnauthorizedEnvelope struct {
	errorEnvelope
	Error struct {
		Code    string `json:"code" example:"unauthorized"`
		Message string `json:"message" example:"unauthorized"`
	} `json:"error"`
}

type ForbiddenEnvelope struct {
	errorEnvelope
	Error struct {
		Code    string `json:"code" example:"forbidden"`
		Message string `json:"message" example:"universal movements can't be edited or deleted"`
	} `json:"error"`
}

type NotFoundEnvelope struct {
	errorEnvelope
	Error struct {
		Code    string `json:"code" example:"not_found"`
		Message string `json:"message" example:"not found"`
	} `json:"error"`
}

type RateLimitedEnvelope struct {
	errorEnvelope
	Error struct {
		Code    string `json:"code" example:"rate_limited"`
		Message string `json:"message" example:"too many requests"`
	} `json:"error"`
}

type InternalErrorEnvelope struct {
	errorEnvelope
	Error struct {
		Code    string `json:"code" example:"internal_error"`
		Message string `json:"message" example:"internal error"`
	} `json:"error"`
}
