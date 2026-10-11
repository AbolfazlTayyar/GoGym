package coach

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newValidationTestService uses a nil *gorm.DB on purpose: input that passes validation panics.
func newValidationTestService() *Service {
	return NewService(NewRepository(nil), "test-secret", time.Hour)
}

func TestCreate_RejectsInvalidInput(t *testing.T) {
	svc := newValidationTestService()

	for name, tc := range map[string]struct {
		input CreateInput
		want  error
	}{
		"international phone": {CreateInput{FirstName: "Ada", LastName: "Lovelace", Phone: "+989372144430", Password: "correct-horse"}, ErrInvalidPhone},
		"short password":      {CreateInput{FirstName: "Ada", LastName: "Lovelace", Phone: "09372144430", Password: "short"}, ErrPasswordTooShort},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.input)
			assert.ErrorIs(t, err, tc.want)
		})
	}
}

func TestResetPassword_RejectsShortPassword(t *testing.T) {
	err := newValidationTestService().ResetPassword(context.Background(), "09372144430", "short")
	assert.ErrorIs(t, err, ErrPasswordTooShort)
}

func TestNewPasswordHash_CountsCharactersNotBytes(t *testing.T) {
	// Seven Persian letters are fourteen bytes but still one character short.
	_, err := newPasswordHash("رمزعبور")
	assert.ErrorIs(t, err, ErrPasswordTooShort)
}
