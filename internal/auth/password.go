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
