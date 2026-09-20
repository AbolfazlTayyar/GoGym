package coach

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := hashPassword("correct-horse-battery-staple")
	require.NoError(t, err)
	assert.NotEqual(t, "correct-horse-battery-staple", hash)

	assert.True(t, verifyPassword(hash, "correct-horse-battery-staple"))
	assert.False(t, verifyPassword(hash, "wrong-password"))
}
