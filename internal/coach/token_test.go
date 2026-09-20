package coach

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueAndParseToken(t *testing.T) {
	coachID := uuid.New()

	token, err := issueToken(coachID, "test-secret", time.Hour)
	require.NoError(t, err)

	got, err := parseToken(token, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, coachID, got)
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := issueToken(uuid.New(), "test-secret", time.Hour)
	require.NoError(t, err)

	_, err = parseToken(token, "wrong-secret")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParseToken_Expired(t *testing.T) {
	token, err := issueToken(uuid.New(), "test-secret", -time.Hour)
	require.NoError(t, err)

	_, err = parseToken(token, "test-secret")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParseToken_Garbage(t *testing.T) {
	_, err := parseToken("not-a-jwt", "test-secret")
	assert.ErrorIs(t, err, ErrInvalidToken)
}
