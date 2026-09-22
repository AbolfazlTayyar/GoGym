package server

import (
	"net/http"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/gin-gonic/gin"
)

// Gin produces a handful of responses that never reach a handler, so nothing
// would put them through the envelope helpers unless it is done here: an
// unmatched path, a path matched with the wrong method, and a panic caught by
// the recovery middleware. Without these three, the only bodies on the API a
// client couldn't parse as an envelope would be exactly the ones it hits when
// something has already gone wrong.

// notFoundHandler answers an unmatched path with an enveloped 404 instead of
// Gin's default plain-text "404 page not found".
func notFoundHandler(c *gin.Context) {
	httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)
}

// methodNotAllowedHandler answers a known path requested with the wrong
// method. It only fires because New sets Engine.HandleMethodNotAllowed —
// Gin's default is to fold this case into the 404.
func methodNotAllowedHandler(c *gin.Context) {
	httpx.Error(c, http.StatusMethodNotAllowed, httpx.CodeMethodNotAllowed, httpx.MsgMethodNotAllowed)
}

// recoveryHandler is the panic response. gin.Recovery()'s own is a bare 500
// with an empty body, which is the one failure a client is least able to
// report usefully: a JSON parse error on an empty string, with no status text
// and no code to branch on. Wrapping it costs these few lines and closes the
// last unenveloped hole in /api/v1, so v1 takes it rather than deferring.
//
// The panic itself is still logged with its stack by gin.CustomRecovery
// before this runs; the client only ever sees the generic message.
func recoveryHandler(c *gin.Context, _ any) {
	httpx.Error(c, http.StatusInternalServerError, httpx.CodeInternalError, httpx.MsgInternalError)
}
