package coach

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidCredentials covers both an unknown phone and a wrong password, so callers can't leak which.
var ErrInvalidCredentials = errors.New("coach: invalid credentials")

type SignupInput struct {
	FirstName string
	LastName  string
	Phone     string
	Password  string
}

type Service struct {
	repo      *Repository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewService(repo *Repository, jwtSecret string, jwtExpiry time.Duration) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

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

func (s *Service) Login(ctx context.Context, phone, password string) (string, error) {
	c, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Compare anyway so an unknown phone isn't measurably faster than a wrong password.
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

func (s *Service) ValidateToken(tokenString string) (uuid.UUID, error) {
	return parseToken(tokenString, s.jwtSecret)
}

// unknownPhoneDummyHash exists only to keep Login's timing consistent for unknown phones.
const unknownPhoneDummyHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5L4pM/AL1CJlXHmoQpEUuC9YCqbmy"
