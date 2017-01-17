package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecaptchaValid(t *testing.T) {

	testCases := []struct {
		response string
		result   bool
		err      bool
	}{
		{`{"success": true, "challenge_ts": 0, "hostname": "localhost"}`,
			true,
			false},
		{`{"success": false, "challenge_ts": 0, "hostname": "localhost"}`,
			false,
			false},
		{`{"success": false, "challenge_ts": 0, "hostname": "localhost",}`,
			false,
			true},
	}

	for _, test := range testCases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, test.response)
		}))
		defer ts.Close()

		url := ts.URL

		r := &Recaptcha{RecaptchaUrl: url, RecaptchaSecret: "test123"}

		result, err := r.Verify("token123")

		if err != nil && test.err == false {
			t.Error("Expected error was not nil")
		} else if err == nil && test.err == true {
			t.Error("Expected error but was nil")
		}

		if result != test.result {
			t.Error("Recaptcha did not parse correct result")
		}

	}
}
