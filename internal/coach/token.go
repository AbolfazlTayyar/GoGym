package coach

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken deliberately hides why a token failed; never surface the reason to clients.
var ErrInvalidToken = errors.New("coach: invalid token")

type claims struct {
	jwt.RegisteredClaims
}

func issueToken(coachID uuid.UUID, secret string, expiry time.Duration) (string, error) {
	now := time.Now()
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   coachID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("coach: failed to sign token: %w", err)
	}
	return signed, nil
}

func parseToken(tokenString, secret string) (uuid.UUID, error) {
	var c claims

	_, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		return uuid.UUID{}, ErrInvalidToken
	}

	coachID, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.UUID{}, ErrInvalidToken
	}

	return coachID, nil
}
