package api

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"io"
)

const SaltLength = 4

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
	Id     int
	User   int
	Method string
	Hash   string
	Salt   string
}
