package httpx_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sort"
	"strings"
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
		// SuccessEnvelope has no Meta field; list routes compose it in, so the runtime adds exactly that key.
		meta := map[string]int{"total": 1}
		assert.ElementsMatch(t,
			keys(t, httpx.Envelope{Success: true, Data: payload{}, Meta: meta}),
			append(keys(t, httpx.SuccessEnvelope{Success: true, Data: payload{}}), "meta"))
	})

	t.Run("no content", func(t *testing.T) {
		c, rec := newTestContext()
		httpx.NoContent(c)

		assert.Equal(t, keys(t, json.RawMessage(rec.Body.Bytes())), keys(t, httpx.NoContentEnvelope{Success: true}))
		assert.JSONEq(t, `{"success":true,"data":null,"error":null}`, rec.Body.String())
	})

	failures := []struct {
		doc    any
		status int
		code   string
		fields map[string]string
	}{
		{httpx.ValidationErrorEnvelope{}, http.StatusBadRequest, httpx.CodeValidationFailed, map[string]string{"name": httpx.MsgFieldRequired}},
		{httpx.InvalidCredentialsEnvelope{}, http.StatusUnauthorized, httpx.CodeInvalidCredentials, nil},
		{httpx.UnauthorizedEnvelope{}, http.StatusUnauthorized, httpx.CodeUnauthorized, nil},
		{httpx.ForbiddenEnvelope{}, http.StatusForbidden, httpx.CodeForbidden, nil},
		{httpx.NotFoundEnvelope{}, http.StatusNotFound, httpx.CodeNotFound, nil},
		{httpx.RateLimitedEnvelope{}, http.StatusTooManyRequests, httpx.CodeRateLimited, nil},
		{httpx.InternalErrorEnvelope{}, http.StatusInternalServerError, httpx.CodeInternalError, nil},
	}
	for _, tc := range failures {
		t.Run(tc.code, func(t *testing.T) {
			c, rec := newTestContext()
			httpx.ErrorFields(c, tc.status, tc.code, "message", tc.fields)

			var body struct {
				Error map[string]json.RawMessage `json:"error"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

			example := docErrorExample(t, tc.doc)
			assert.Equal(t, tc.code, example["code"])
			assert.Equal(t, keys(t, body.Error), keys(t, example))
			assert.Equal(t, keys(t, json.RawMessage(rec.Body.Bytes())), keys(t, tc.doc))
		})
	}
}

// docErrorExample maps a doc envelope's error fields to their example tags, since swag reads examples from there.
func docErrorExample(t *testing.T, doc any) map[string]string {
	t.Helper()

	errField, ok := reflect.TypeOf(doc).FieldByName("Error")
	require.True(t, ok)

	out := make(map[string]string, errField.Type.NumField())
	for i := range errField.Type.NumField() {
		f := errField.Type.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		out[name] = f.Tag.Get("example")
	}

	return out
}
