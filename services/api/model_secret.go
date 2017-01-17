package api

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

const SecretLength = 32

// Global secret generator using cryptographically secure random number
// generator
var Secret = &SecretGenerator{rand.Reader}

type SecretGenerator struct {
	RandReader io.Reader
}

// Given a password, generate a salt and hash it and return the salt and
// hash and any error encountered.
func (gen *SecretGenerator) Generate() (string, error) {
	raw := make([]byte, SecretLength)
	_, err := io.ReadFull(gen.RandReader, raw)
	if err != nil {
		return "", err
	}

	secret := make([]byte, base64.RawURLEncoding.EncodedLen(len(raw)))
	base64.RawURLEncoding.Encode(secret, raw)

	return string(secret), nil
}
