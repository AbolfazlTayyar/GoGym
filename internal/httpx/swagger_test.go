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

// TestEnvelopeDocsMatchRuntime is the only guard against the hand-written doc envelopes drifting.
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
		// ErrorEnvelope reuses ErrorBody, so only the helper's output needs checking.
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
