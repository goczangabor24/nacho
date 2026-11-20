package auth

import (
	"errors"
	"fmt"

	"github.com/alexedwards/argon2id"
)

const (
	passwordMinLen = 10
	passwordMaxLen = 128
)

var (
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrInvalidHash        = errors.New("auth: malformed password hash")
)

var hashParams = &argon2id.Params{
	Memory:      19456,
	Iterations:  2,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

func ValidatePassword(password string) error {
	if len(password) < passwordMinLen {
		return fmt.Errorf("auth: password must be at least %d characters", passwordMinLen)
	}
	if len(password) > passwordMaxLen {
		return fmt.Errorf("auth: password must be at most %d characters", passwordMaxLen)
	}
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); != nil {
		return "", err
	}
	hash, err := argon2id.CreateHash(password, hashParams)
	if err != nil {
		return "", fmt.Errorf("auth: error hashing password: %w", err)
	}
	return hash, nil
}

func ComparePassword(hash, password string) error {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidHash, err)
	}
}

var dummyHash = mustHash("a-password-that-is-never-valid-like-this-one-here")

func mustHash(password string) string {
	hash, err := HashPassword(password)
	if err != nil {
		panic("auth: cannot hash dummy password: " + err.Error())
	}
	return hash
}
