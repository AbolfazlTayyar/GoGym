package httpx

// The two types below exist only for swagger annotations. Handlers always
// write the single Envelope type; these describe its two occupied states so
// Swagger UI can render an honest example of each.
//
// One type can't do that job. Envelope.Error is a *ErrorBody, and a schema
// has no way to say "this field is populated on failure and null on success"
// — so a single annotated type renders every success example with a filled-in
// error object inside it, and every failure example with "success": true.
// Splitting the documented shape in two is the "explicit envelope structs"
// fallback the API envelope task called for, applied to the success/failure
// split rather than per payload, so payloads still compose with {data=...}.
//
// envelopeDocsMatchRuntime in swagger_test.go fails if these drift from the
// keys Envelope actually marshals.

// SuccessEnvelope documents a success response. Annotate the payload onto it
// with the swaggo composition syntax:
//
//	@Success 200 {object} httpx.SuccessEnvelope{data=coachResponse}
type SuccessEnvelope struct {
	// Success is always true on these responses.
	Success bool `json:"success" example:"true"`
	// Data is the payload, replaced by the annotation's {data=...} override.
	Data any `json:"data"`
	// Error is always null on a success response.
	Error any `json:"error"`
	// Meta carries list metadata such as pagination, and is omitted from
	// the body entirely when empty — which is every endpoint today.
	Meta any `json:"meta,omitempty"`
}

// ErrorEnvelope documents a failure response:
//
//	@Failure 400 {object} httpx.ErrorEnvelope
type ErrorEnvelope struct {
	// Success is always false on these responses.
	Success bool `json:"success" example:"false"`
	// Data is always null on a failure response.
	Data any `json:"data"`
	// Error is the failure detail.
	Error ErrorBody `json:"error"`
}
