package coach

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken is returned by parseToken for any token that fails
// signature, expiry, or claim validation. Callers must not distinguish the
// underlying reason in a response — that would leak information useful to
// an attacker probing for valid coach ids or token formats.
var ErrInvalidToken = errors.New("coach: invalid token")

// claims are the JWT claims issued at login. Subject carries the coach id;
// expiry comes from config.Config.JWTExpiry at issuance time.
type claims struct {
	jwt.RegisteredClaims
}

// issueToken signs a JWT for coachID, valid for expiry, using secret.
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

// parseToken validates tokenString's signature and expiry against secret
// and returns the coach id from its subject claim, or ErrInvalidToken.
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
