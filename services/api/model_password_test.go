package api

import (
	"bytes"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"testing"
)

func TestHashPassword(t *testing.T) {

	hasher := &PasswordHasher{bytes.NewReader([]byte("Hello, World!"))}

	testCases := []struct {
		passw    string
		expected *Password
		err      error
	}{
		{"t3stP455",
			&Password{
				0, "bcrypt", "$2a$10$kB1KL3oXK0/nwbPr0lNfEejixlQNQoTkHzmt.zWlB4DNCSAvpP0pW", "SGVsbA",
			},
			nil,
		},
	}

	for _, test := range testCases {
		pw, err := hasher.hash(test.passw)

		if !reflect.DeepEqual(err, test.err) {
			t.Error("Expected error does not match.")
		}

		if pw.Method != test.expected.Method {
			t.Error("Password hashing method is different.")
		}

		if pw.Salt != test.expected.Salt {
			t.Error("Password salt is different.")
		}

		combined := append([]byte(pw.Salt), test.passw...)
		if pw.Method == "bcrypt" && bcrypt.CompareHashAndPassword([]byte(pw.Hash), combined) != nil {
			t.Error("Password failed comparison.")
		}
	}
}
