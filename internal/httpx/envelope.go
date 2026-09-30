// Package httpx writes the /api/v1 response envelope; handlers must use its helpers, never c.JSON.
package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope's Data and Error deliberately lack omitempty so both keys are always present.
type Envelope struct {
	Success bool       `json:"success" example:"true"`
	Data    any        `json:"data"`
	Error   *ErrorBody `json:"error"`
	Meta    any        `json:"meta,omitempty"`
}

type ErrorBody struct {
	// Code is one of the Code* constants; clients branch on it, unlike Message.
	Code string `json:"code" example:"validation_failed"`
	// Message is human-readable and may change freely.
	Message string `json:"message" example:"invalid request"`
	// Fields is set for request-validation failures only.
	Fields map[string]string `json:"fields,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

func OKWithMeta(c *gin.Context, data, meta any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

// NoContent sends a 200 with null data, since a real 204 has no body to parse as an Envelope.
func NoContent(c *gin.Context) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: nil})
}

// Error aborts the chain so it is also safe to use from middleware.
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, Envelope{Success: false, Error: &ErrorBody{Code: code, Message: message}})
}

// ErrorFields omits a nil or empty fields map, so ValidationFields' result can be passed as is.
func ErrorFields(c *gin.Context, status int, code, message string, fields map[string]string) {
	c.AbortWithStatusJSON(status, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message, Fields: fields},
	})
}
