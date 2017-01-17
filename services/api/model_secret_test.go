package api

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestGenerateSecret(t *testing.T) {

	testCases := []struct {
		gen      *SecretGenerator
		expected string
		err      error
	}{
		{
			&SecretGenerator{bytes.NewReader([]byte("abcdefghijklmnopqrstuvwxyz1234567890"))},
			"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY",
			nil,
		},
		{
			&SecretGenerator{bytes.NewReader([]byte("Hello, World!"))},
			"",
			io.ErrUnexpectedEOF,
		},
	}

	for _, test := range testCases {
		secret, err := test.gen.Generate()

		fmt.Println(secret)

		if !reflect.DeepEqual(err, test.err) {
			fmt.Println(err)
			t.Error("Expected error does not match.")
		}

		if test.expected != secret {
			t.Error("Secret did not match expectation.")
		}
	}
}
