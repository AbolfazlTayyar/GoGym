package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PathUUID answers a malformed id with 404, not 400: no row can have it, so it simply doesn't exist.
func PathUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		Error(c, http.StatusNotFound, CodeNotFound, MsgNotFound)
		return uuid.Nil, false
	}

	return id, true
}
