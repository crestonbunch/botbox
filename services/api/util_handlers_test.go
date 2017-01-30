package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/gorilla/mux"
)

func TestJsonHandler(t *testing.T) {
	type Model struct {
		Dummy string `json:"dummy"`
	}

	type sampleResponse struct {
		Model *Model
		Error error
	}

	testCases := map[*http.Request]sampleResponse{
		// Good case
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "valid"
		}`)): sampleResponse{
			Model: &Model{
				Dummy: "valid",
			},
			Error: nil,
		},

		// Invalid JSON
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "invalid",
		}`)): sampleResponse{Model: nil, Error: ErrInvalidJson},

		// Invalid reader
		httptest.NewRequest(
			"POST", "http://local/user/password",
			iotest.TimeoutReader(strings.NewReader(`{
					"dummy": "invalid"
				}`)),
		): sampleResponse{Model: nil, Error: ErrUnknown},
	}

	for req, expected := range testCases {
		handler := &JsonHandler{Target: func() interface{} { return &Model{} }}
		model, err := handler.Handle(req)

		if err != nil && expected.Error == nil {
			t.Error(err)
		} else if err == nil && expected.Error != nil {
			t.Error("Post model returned a nil error!")
		} else if expected.Error == nil && err == nil {
			// Why does deep equal not correctly compare nil and nil?
		} else if !reflect.DeepEqual(expected.Error, err) {
			t.Error("Post model returned the wrong error!")
		}

		if model != nil {
			if !reflect.DeepEqual(*expected.Model, *model.(*Model)) {
				t.Error("Model loaded the wrong data!")
			}
		}
	}
}

type mockSession struct {
	ReturnUser       int
	ReturnUserError  error
	ReturnPerms      PermissionSet
	ReturnPermsError error
}

func (s *mockSession) GetUserId(string) (int, error) {
	return s.ReturnUser, s.ReturnUserError
}

func (s *mockSession) GetPermissions(string) (PermissionSet, error) {
	return s.ReturnPerms, s.ReturnPermsError
}

func TestJsonHandlerWithId(t *testing.T) {
	type Model struct {
		Dummy string `json:"dummy"`
	}

	type sampleResponse struct {
		Auth  string
		Sess  *mockSession
		Model *Model
		Error error
		User  int
	}

	testCases := map[*http.Request]sampleResponse{
		// Good case
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "valid"
		}`)): sampleResponse{
			Auth: "Bearer 12345",
			Sess: &mockSession{ReturnUser: 101},
			Model: &Model{
				Dummy: "valid",
			},
			Error: nil,
			User:  101,
		},

		// Invalid JSON
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "invalid",
		}`)): sampleResponse{
			Auth:  "Bearer 12345",
			Sess:  nil,
			Model: nil,
			Error: ErrInvalidJson,
		},

		// Invalid reader
		httptest.NewRequest(
			"POST", "http://local/user/password",
			iotest.TimeoutReader(strings.NewReader(`{
					"dummy": "invalid"
				}`)),
		): sampleResponse{
			Auth:  "Bearer 12345",
			Sess:  nil,
			Model: nil,
			Error: ErrUnknown,
		},

		// Session error
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "valid"
		}`)): sampleResponse{
			Auth: "Bearer 12345",
			Sess: &mockSession{ReturnUserError: errors.New("Dummy error")},
			Model: &Model{
				Dummy: "valid",
			},
			Error: ErrUnknown,
		},
	}

	for req, expected := range testCases {
		req.Header.Set(AuthorizationHeader, expected.Auth)
		handler := &JsonHandlerWithAuth{
			Target:  func() interface{} { return &Model{} },
			session: expected.Sess,
		}
		model, err := handler.HandleWithId(req)

		if err != nil && expected.Error == nil {
			t.Error(err)
		} else if err == nil && expected.Error != nil {
			t.Error("Post model returned a nil error!")
		} else if expected.Error == nil && err == nil {
			// Why does deep equal not correctly compare nil and nil?
		} else if !reflect.DeepEqual(expected.Error, err) {
			t.Error("Post model returned the wrong error!")
		}

		if handler.User != expected.User {
			t.Error("User was not set correctly.")
		}

		if expected.Model != nil && model != nil &&
			!reflect.DeepEqual(*expected.Model, *model.(*Model)) {
			t.Error("Model loaded the wrong data!")
		}
	}
}

func TestJsonHandlerWithPermissions(t *testing.T) {
	type Model struct {
		Dummy string `json:"dummy"`
	}

	type sampleResponse struct {
		Auth  string
		Sess  *mockSession
		Model *Model
		Error error
		User  int
		Perms PermissionSet
	}

	testCases := map[*http.Request]sampleResponse{
		// Good case
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "valid"
		}`)): sampleResponse{
			Auth: "Bearer 12345",
			Sess: &mockSession{
				ReturnUser:  101,
				ReturnPerms: PermissionSet([]string{"DUMMY_PERM"}),
			},
			Model: &Model{
				Dummy: "valid",
			},
			Error: nil,
			User:  101,
			Perms: PermissionSet([]string{"DUMMY_PERM"}),
		},

		// Invalid JSON
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "invalid",
		}`)): sampleResponse{
			Auth:  "Bearer 12345",
			Sess:  nil,
			Model: nil,
			Error: ErrInvalidJson,
		},

		// Invalid reader
		httptest.NewRequest(
			"POST", "http://local/user/password",
			iotest.TimeoutReader(strings.NewReader(`{
					"dummy": "invalid"
				}`)),
		): sampleResponse{
			Auth:  "Bearer 12345",
			Sess:  nil,
			Model: nil,
			Error: ErrUnknown,
		},

		// User error
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "valid"
		}`)): sampleResponse{
			Auth: "Bearer 12345",
			Sess: &mockSession{ReturnUserError: errors.New("Dummy error")},
			Model: &Model{
				Dummy: "valid",
			},
			Error: ErrUnknown,
		},

		// Perms error
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"dummy": "valid"
		}`)): sampleResponse{
			Auth: "Bearer 12345",
			Sess: &mockSession{
				ReturnUser:       101,
				ReturnPermsError: errors.New("Dummy error"),
			},
			Model: &Model{
				Dummy: "valid",
			},
			User:  101,
			Error: ErrUnknown,
		},
	}

	for req, expected := range testCases {
		req.Header.Set(AuthorizationHeader, expected.Auth)
		handler := &JsonHandlerWithAuth{
			Target:  func() interface{} { return &Model{} },
			session: expected.Sess,
		}
		model, err := handler.HandleWithPermissions(req)

		if err != nil && expected.Error == nil {
			t.Error(err)
		} else if err == nil && expected.Error != nil {
			t.Error("Post model returned a nil error!")
		} else if expected.Error == nil && err == nil {
			// Why does deep equal not correctly compare nil and nil?
		} else if !reflect.DeepEqual(expected.Error, err) {
			t.Error("Post model returned the wrong error!")
		}

		if handler.User != expected.User {
			t.Errorf("User %d was not set correctly, expected %d.",
				handler.User, expected.User)
		}
		if !reflect.DeepEqual(expected.Perms, handler.Permissions) {
			t.Errorf("Permissions %+v were not set correctly, expected %+v.",
				handler.Permissions, expected.Perms)
		}

		if expected.Model != nil && model != nil &&
			!reflect.DeepEqual(*expected.Model, *model.(*Model)) {
			t.Error("Model loaded the wrong data!")
		}
	}
}

func TestURLPathHandler(t *testing.T) {
	type Model struct {
		Dummy string `json:"dummy"`
	}

	type sampleResponse struct {
		Model *Model
		Error *HttpError
	}

	testCases := map[string]sampleResponse{
		// Good case
		"/path/var1": sampleResponse{
			Model: &Model{
				Dummy: "var1",
			},
			Error: nil,
		},
	}

	for uri, expected := range testCases {
		handler := &URLPathHandler{Target: func() interface{} { return &Model{} }}

		r := mux.NewRouter()
		r.HandleFunc("/path/{dummy}", func(w http.ResponseWriter, req *http.Request) {
			model, err := handler.Handle(req)

			if err != expected.Error {
				t.Error("Error was not expected.")
			}

			if model != nil {
				if !reflect.DeepEqual(*expected.Model, *model.(*Model)) {
					t.Error("Model loaded the wrong data!")
				}
			}
		})

		serv := httptest.NewServer(r)
		defer serv.Close()

		_, err := http.Get(serv.URL + uri)
		if err != nil {
			t.Error(err)
		}
	}
}
