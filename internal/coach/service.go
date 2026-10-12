package coach

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/AbolfazlTayyar/gogym/internal/platform/validate"
	"github.com/google/uuid"
)

const minPasswordLength = 8

// ErrInvalidCredentials covers both an unknown phone and a wrong password, so callers can't leak which.
var ErrInvalidCredentials = errors.New("coach: invalid credentials")

var (
	ErrInvalidPhone     = errors.New("coach: phone " + validate.MsgIranMobile)
	ErrPasswordTooShort = fmt.Errorf("coach: password must be at least %d characters", minPasswordLength)
)

type CreateInput struct {
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

func (s *Service) Create(ctx context.Context, input CreateInput) (*Coach, error) {
	if !validate.IsIranMobile(input.Phone) {
		return nil, ErrInvalidPhone
	}

	hash, err := newPasswordHash(input.Password)
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

// ResetPassword doesn't end existing sessions: tokens already issued stay valid until they expire.
func (s *Service) ResetPassword(ctx context.Context, phone, password string) error {
	hash, err := newPasswordHash(password)
	if err != nil {
		return err
	}

	return s.repo.UpdatePasswordHash(ctx, phone, hash)
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

func newPasswordHash(password string) (string, error) {
	// Counts characters, not bytes, so the rule reads the same for a Persian password.
	if utf8.RuneCountInString(password) < minPasswordLength {
		return "", ErrPasswordTooShort
	}
	return hashPassword(password)
}

// unknownPhoneDummyHash exists only to keep Login's timing consistent for unknown phones.
const unknownPhoneDummyHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5L4pM/AL1CJlXHmoQpEUuC9YCqbmy"
