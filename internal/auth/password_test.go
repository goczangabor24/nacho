package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/alexedwards/argon2id"
)

func TestHashAndCompare(t *testing.T) {
	hash, err := HashPassword("correct horse battery")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := ComparePassword(hash, "correct horse battery"); err != nil {
		t.Fatalf("compare with correct password: %v", err)
	}
	if err := ComparePassword(hash, "wrong password here"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestHashIsSaltedAndFormatted(t *testing.T) {
	const password = "corect horse battery"

	first, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	second, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	if first == second {
		t.Fatal("two hashes of the same password must differ")
	}
	if !strings.HasPrefix(first, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected prefix: %q", first)
	}
	if strings.Contains(first, password) {
		t.Fatal("hash must not contain the plaintext")
	}
	if err := ComparePassword(second, password); err != nil {
		t.Fatalf("second hash must verify: %v", err)
	}
}

func TestValidatePasswordLengths(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{"too short", 9, true},
		{"minimum", 10, false},
		{"maximum", 128, false},
		{"too long", 129, true},
	}

	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			err := ValidatePassword(strings.Repeat("a", testcase.length))
			if (err != nil) != testcase.wantErr {
				t.Fatalf("length %d: want error %v, got %v", testcase.length, testcase.wantErr, err)
			}
		})
	}
}

func TestMalformedHashes(t *testing.T) {
	hashes := []string{
		"",
		"garbage",
		"$bcrypt$v=19$m=19456,t=2,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=abc,t=2,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=19456,t=2$c2FsdA$a2V5",
		"$argon2id$v=19$m=19456,t=2,p=1$!!!$a2V5",
		"$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$!!!",
		"$argon2id$v=19$m=99999999,t=2,p=1$c2FsdA$a2V5",
	}

	for _, hash := range hashes {
		t.Run(hash, func(t *testing.T) {
			err := ComparePassword(hash, "correct horse battery")
			if !errors.Is(err, ErrInvalidHash) {
				t.Fatalf("want ErrInvalidHash, got %v", err)
			}
			if errors.Is(err, ErrInvalidCredentials) {
				t.Fatal("malformed hash must not look like a wrong password")
			}
		})
	}
}

func TestCompareUsesStoredParams(t *testing.T) {
	hash, err := argon2id.CreateHash("correct horse battery", &argon2id.Params{
		Memory:      9216,
		Iterations:  4,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		t.Fatalf("create hash: %v", err)
	}
	if err := ComparePassword(hash, "correct horse battery"); err != nil {
		t.Fatalf("hash with non-default params must verify: %v", err)
	}
}
