package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestEndpoint() *Endpoint {
	return &Endpoint{
		Path:    "/",
		Methods: []string{"GET", "POST"},
		Handler: testHandler,
		Processors: []Processor{
			testProcessor,
		},
		Writer: testWriter,
	}
}

func testHandler(r *http.Request) (interface{}, *HttpError) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, ErrUnknown
	}

	if string(body) == "failHandler" {
		return nil, ErrUnknown
	} else {
		return string(body), nil
	}
}

func testProcessor(input interface{}) (interface{}, *HttpError) {
	if input.(string) == "failProcessor" {
		return nil, ErrUnknown
	}
	return input.(string) + "!", nil
}

func testWriter(input interface{}) ([]byte, *HttpError) {
	if input.(string) == "failWriter!" {
		return nil, ErrUnknown
	}
	return []byte(input.(string)), nil
}

func TestApp(t *testing.T) {
	app := NewApp(nil, nil, nil)
	app.Attach(newTestEndpoint())

	testCases := []struct {
		body   string
		output string
	}{
		{
			body:   "Hello, World",
			output: "Hello, World!",
		},
		{
			body:   "failHandler",
			output: ErrUnknown.Error(),
		},
		{
			body:   "failProcessor",
			output: ErrUnknown.Error(),
		},
		{
			body:   "failWriter",
			output: ErrUnknown.Error(),
		},
	}

	srv := httptest.NewServer(app.router)
	defer srv.Close()

	for _, test := range testCases {
		res, err := http.Post(srv.URL, "text/plain", bytes.NewReader([]byte(test.body)))

		if err != nil {
			t.Error(err)
		}
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Error(err)
		}
		if strings.TrimSpace(string(body)) != test.output {
			t.Error("Output does not match")
		}
	}
}
