package httpx_test

import (
	"encoding/json"
	"net/http"
	"sort"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// keys returns the top-level JSON keys v marshals to, sorted.
func keys(t *testing.T, v any) []string {
	t.Helper()

	encoded, err := json.Marshal(v)
	require.NoError(t, err)

	var decoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &decoded))

	out := make([]string, 0, len(decoded))
	for k := range decoded {
		out = append(out, k)
	}
	sort.Strings(out)

	return out
}

// TestEnvelopeDocsMatchRuntime is the guard the swagger.go comment points at:
// SuccessEnvelope and ErrorEnvelope are written by hand for the annotations,
// so nothing but this test stops them documenting a body that Envelope no
// longer produces.
func TestEnvelopeDocsMatchRuntime(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.Equal(t,
			keys(t, httpx.Envelope{Success: true, Data: payload{}}),
			keys(t, httpx.SuccessEnvelope{Success: true, Data: payload{}}))
	})

	t.Run("success with meta", func(t *testing.T) {
		meta := map[string]int{"total": 1}
		assert.Equal(t,
			keys(t, httpx.Envelope{Success: true, Data: payload{}, Meta: meta}),
			keys(t, httpx.SuccessEnvelope{Success: true, Data: payload{}, Meta: meta}))
	})

	t.Run("failure", func(t *testing.T) {
		errBody := httpx.ErrorBody{Code: httpx.CodeNotFound, Message: httpx.MsgNotFound}
		assert.Equal(t,
			keys(t, httpx.Envelope{Error: &errBody}),
			keys(t, httpx.ErrorEnvelope{Error: errBody}))
	})

	t.Run("error body", func(t *testing.T) {
		// The documented error detail is the same type the helpers write,
		// so this only has to hold for the helper's own output.
		c, rec := newTestContext()
		httpx.Error(c, http.StatusNotFound, httpx.CodeNotFound, httpx.MsgNotFound)

		var body struct {
			Error json.RawMessage `json:"error"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

		var errBody httpx.ErrorBody
		require.NoError(t, json.Unmarshal(body.Error, &errBody))
		assert.Equal(t, httpx.CodeNotFound, errBody.Code)
	})
}
