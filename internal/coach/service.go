package coach

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidCredentials is returned by Login when the phone/password pair
// doesn't match — deliberately the same error for "no such phone" and
// "wrong password" so handlers can't leak which one it was.
var ErrInvalidCredentials = errors.New("coach: invalid credentials")

// SignupInput is the data needed to create a coach account.
type SignupInput struct {
	FirstName string
	LastName  string
	Phone     string
	Password  string
}

// Service holds the coach module's business logic: signup, login, and the
// JWT issuance/validation the auth middleware relies on.
type Service struct {
	repo      *Repository
	jwtSecret string
	jwtExpiry time.Duration
}

// NewService builds a Service. jwtSecret and jwtExpiry come from
// config.Config (JWTSecret, JWTExpiry).
func NewService(repo *Repository, jwtSecret string, jwtExpiry time.Duration) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

// Signup hashes input.Password with bcrypt and creates the coach. The
// plaintext password is never stored or logged past this call.
func (s *Service) Signup(ctx context.Context, input SignupInput) (*Coach, error) {
	hash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	c := &Coach{
		ID:           uuid.New(),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Phone:        input.Phone,
		PasswordHash: hash,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

// Login verifies phone/password against the stored bcrypt hash and, on
// success, returns a signed JWT. It returns ErrInvalidCredentials for both
// an unknown phone and a wrong password.
func (s *Service) Login(ctx context.Context, phone, password string) (string, error) {
	c, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Still run a bcrypt comparison against a fixed hash so a
			// nonexistent phone doesn't respond measurably faster than a
			// wrong password would, which would itself leak which case hit.
			verifyPassword(unknownPhoneDummyHash, password)
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !verifyPassword(c.PasswordHash, password) {
		return "", ErrInvalidCredentials
	}

	token, err := issueToken(c.ID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken parses and validates tokenString, returning the coach id it
// was issued for. Used by the auth middleware.
func (s *Service) ValidateToken(tokenString string) (uuid.UUID, error) {
	return parseToken(tokenString, s.jwtSecret)
}

// unknownPhoneDummyHash is a bcrypt hash of an arbitrary fixed string, used
// only to keep Login's timing consistent when no coach matches the phone.
const unknownPhoneDummyHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5L4pM/AL1CJlXHmoQpEUuC9YCqbmy"
