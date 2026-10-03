package httpx_test

import (
	"net/http"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathUUID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		want := uuid.New()
		c, rec := newTestContext()
		c.Params = gin.Params{{Key: "id", Value: want.String()}}

		got, ok := httpx.PathUUID(c, "id")

		require.True(t, ok)
		assert.Equal(t, want, got)
		assert.False(t, c.IsAborted())
		assert.Empty(t, rec.Body.String(), "nothing is written on success")
	})

	for name, raw := range map[string]string{
		"malformed": "not-a-uuid",
		"missing":   "",
	} {
		t.Run(name, func(t *testing.T) {
			c, rec := newTestContext()
			c.Params = gin.Params{{Key: "id", Value: raw}}

			_, ok := httpx.PathUUID(c, "id")

			require.False(t, ok)
			assert.True(t, c.IsAborted())
			assert.Equal(t, http.StatusNotFound, rec.Code)

			errBody, isMap := decode(t, rec)["error"].(map[string]any)
			require.True(t, isMap)
			assert.Equal(t, httpx.CodeNotFound, errBody["code"])
		})
	}
}
