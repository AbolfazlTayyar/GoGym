// Package httpx defines the single response envelope every /api/v1 endpoint
// returns, and the helpers that write it.
//
// Every response body — success or failure — is an Envelope: a boolean
// success, a data payload, an error object, and an optional meta. Exactly one
// of data/error is non-null; both keys are always present so the frontend can
// destructure a response without first branching on the status code.
//
// Enforcement is by convention, not by middleware: handlers call these helpers
// and never call c.JSON / c.AbortWithStatusJSON themselves. A response-rewriting
// middleware that wrapped whatever a handler wrote was considered and rejected —
// it has to buffer every response body to re-marshal it, and it moves the real
// shape of the response out of the handler, so the swagger annotation on the
// handler (which documents the bare payload) silently stops describing what the
// endpoint actually returns. Explicit helpers keep the annotation and the body
// written in the same place.
//
// /healthz is deliberately exempt — see internal/server/healthz.go.
package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the body of every /api/v1 response.
//
// Success and Error are rendered even when nil (no omitempty) so the two keys
// are always present; Meta is omitted when empty.
type Envelope struct {
	Success bool       `json:"success" example:"true"`
	Data    any        `json:"data"`
	Error   *ErrorBody `json:"error"`
	// Meta carries list metadata such as pagination. Nothing populates it
	// yet — it is defined here so the endpoints that need it later add a
	// key rather than change the envelope.
	Meta any `json:"meta,omitempty"`
}

// ErrorBody is the error detail of a failed response.
type ErrorBody struct {
	// Code is a stable machine-readable identifier the frontend branches
	// on. It is one of the Code* constants and does not change wording
	// with the message.
	Code string `json:"code" example:"validation_failed"`
	// Message is human-readable and may change freely.
	Message string `json:"message" example:"invalid request"`
	// Fields maps a request field name to why it was rejected. It is set
	// for request-validation failures only, and omitted otherwise.
	Fields map[string]string `json:"fields,omitempty"`
}

// OK writes a 200 with data as the payload.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

// OKWithMeta writes a 200 with data as the payload and meta alongside it, for
// list endpoints that report pagination.
func OKWithMeta(c *gin.Context, data, meta any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

// Created writes a 201 with data as the payload.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

// NoContent writes the envelope's stand-in for a 204: a 200 whose data is
// null. A real 204 carries no body by definition, which would make it the one
// success response on the API a client couldn't parse as an Envelope.
func NoContent(c *gin.Context) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: nil})
}

// Error writes a failure envelope with the given status, machine-readable
// code and human-readable message.
//
// It aborts the handler chain so the same helper works from middleware (where
// not aborting would let the request continue to the handler it just rejected)
// as from a handler, where there is nothing left to abort.
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, Envelope{Success: false, Error: &ErrorBody{Code: code, Message: message}})
}

// ErrorFields is Error with a field name -> message map attached, for
// request-validation failures. A nil or empty fields map is omitted from the
// body, making this safe to call with the result of ValidationFields.
func ErrorFields(c *gin.Context, status int, code, message string, fields map[string]string) {
	c.AbortWithStatusJSON(status, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message, Fields: fields},
	})
}
