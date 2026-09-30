package httpx

// Swagger-only types: one Envelope schema can't say "error is null on success", so each state gets its own.

// SuccessEnvelope takes the payload via swag composition: httpx.SuccessEnvelope{data=...}.
type SuccessEnvelope struct {
	Success bool `json:"success" example:"true"`
	Data    any  `json:"data"`
	Error   any  `json:"error"`
	Meta    any  `json:"meta,omitempty"`
}

type ErrorEnvelope struct {
	Success bool      `json:"success" example:"false"`
	Data    any       `json:"data"`
	Error   ErrorBody `json:"error"`
}
