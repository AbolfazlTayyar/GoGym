package server

import (
	"net/http"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/gin-gonic/gin"
)

// Gin answers unmatched paths, wrong methods, and panics without a handler; these keep them enveloped.

func notFoundHandler(c *gin.Context) {
	httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
}

// methodNotAllowedHandler only fires because New sets Engine.HandleMethodNotAllowed.
func methodNotAllowedHandler(c *gin.Context) {
	httpx.Error(c, http.StatusMethodNotAllowed, httpx.CodeMethodNotAllowed, httpx.MsgMethodNotAllowed)
}

// recoveryHandler replaces gin.Recovery's empty-body 500; gin.CustomRecovery still logs the stack.
func recoveryHandler(c *gin.Context, _ any) {
	httpx.Error(c, http.StatusInternalServerError, httpx.CodeInternalError, httpx.MsgInternalError)
}
