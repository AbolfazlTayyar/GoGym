package membership

import (
	"errors"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisitDate(t *testing.T) {
	start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	// 01:00 in Tehran on 17 Nov is still 16 Nov in UTC: the phone's today is the server's tomorrow.
	now := time.Date(2026, 11, 16, 21, 30, 0, 0, time.UTC)

	valid := []string{"2026-11-01", "2026-11-16", "2026-11-17", " 2026-11-10 "}
	for _, raw := range valid {
		t.Run("accepts "+raw, func(t *testing.T) {
			_, err := visitDate(raw, start, now)
			assert.NoError(t, err)
		})
	}

	invalid := map[string]string{
		"":           httpx.MsgFieldRequired,
		"16/11/2026": httpx.MsgDateFormat,
		"2026-10-31": "must be between 2026-11-01 and 2026-11-17",
		"2026-11-18": "must be between 2026-11-01 and 2026-11-17",
	}
	for raw, want := range invalid {
		t.Run("rejects "+raw, func(t *testing.T) {
			_, err := visitDate(raw, start, now)

			var verr *ValidationError
			require.True(t, errors.As(err, &verr))
			assert.Equal(t, want, verr.Fields[fieldDate])
		})
	}
}
