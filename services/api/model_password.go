package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
)

const SaltLength = 4
const MinPasswordLen = 6

// Global password hasher using cryptographically secure random number generator
var Hasher = &PasswordHasher{rand.Reader}

type PasswordHasher struct {
	RandReader io.Reader
}

// Given a password, generate a salt and hash it and return the salt and
// hash and any error encountered.
func (hasher *PasswordHasher) hash(passw string) (*Password, error) {
	raw := make([]byte, SaltLength)
	_, err := io.ReadFull(hasher.RandReader, raw)
	if err != nil {
		return nil, err
	}

	salt := make([]byte, base64.RawURLEncoding.EncodedLen(len(raw)))
	base64.RawURLEncoding.Encode(salt, raw)

	combined := append(salt, passw...)
	hash, err := bcrypt.GenerateFromPassword(combined, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Password{
		Salt:   string(salt),
		Hash:   string(hash),
		Method: "bcrypt",
	}, nil
}

type Password struct {
	User   int    `db:"user"`
	Method string `db:"method"`
	Hash   string `db:"hash"`
	Salt   string `db:"salt"`
}

func (p *Password) Matches(other string) error {
	switch p.Method {
	case "bcrypt":
		combined := p.Salt + other
		return bcrypt.CompareHashAndPassword([]byte(p.Hash), []byte(combined))
	default:
		return errors.New("Unsupported password type.")
	}
}

// Make sure a password is secure enough
func ValidatePassword(password string) *HttpError {
	if password == "" {
		return ErrMissingPassword
	}
	if len(password) < MinPasswordLen {
		return ErrPasswordTooShort
	}

	return nil
}
