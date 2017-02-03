package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNullTimeMarshal(t *testing.T) {
	tm := time.Now()
	e, _ := json.Marshal(tm)
	tests := []struct {
		Time   *NullTime
		Expect []byte
	}{
		{Time: &NullTime{Valid: false}, Expect: []byte("null")},
		{Time: &NullTime{Time: tm, Valid: true}, Expect: e},
	}

	for _, test := range tests {
		expected := test.Expect
		actual, err := json.Marshal(test.Time)

		if err != nil {
			t.Error(err)
		}

		if string(expected) != string(actual) {
			t.Errorf("Time was %s not %s\n", actual, expected)
		}
	}
}
