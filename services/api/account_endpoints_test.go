package api

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

type mockSender struct{}

func (s *mockSender) SendVerificationEmail(to string, secret string) error {
	return nil
}
func (s *mockSender) SendPasswordRecoveryEmail(to string, secret string) error {
	return nil
}

type mockChecker struct{}

func (c *mockChecker) IsHuman(secret string) (bool, error) {
	if secret == "12345" {
		return true, nil
	}
	return false, nil
}

func TestAccountNew(t *testing.T) {
	userModel := newMockUserModel()
	accountEndpoints := &AccountEndpoints{
		userModel, &mockSender{}, &mockChecker{},
	}

	userModel.ExpectSelectByUsername().WithArgs("user")
	userModel.ExpectSelectByEmail().WithArgs("email@example.com")
	userModel.ExpectInsert().WithArgs("user", "email@example.com", "password").
		WillReturn(1)
	userModel.ExpectCreateVerificationSecret().WithArgs(1, "email@example.com")

	buf := bytes.NewBuffer([]byte(`{
		"username": "user",
		"email": "email@example.com",
		"password": "password",
		"captcha": "12345"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	accountEndpoints.AccountNew(w, req)

	if w.Result().StatusCode != 200 {
		t.Error("New account did not return 200")
	}

	if err := userModel.ExpectationsWereMetInOrder(); err != nil {
		t.Error(err)
	}
}

func TestAccountNewInvalidBody(t *testing.T) {
	userModel := newMockUserModel()
	accountEndpoints := &AccountEndpoints{
		userModel, &mockSender{}, &mockChecker{},
	}

	buf := bytes.NewBuffer([]byte(`{
		"username": "user",
		"email": "email@example.com",
		"password": "password",
		"captcha": "12345",
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	accountEndpoints.AccountNew(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("New account did not return 400")
	}

	if err := userModel.ExpectationsWereMetInOrder(); err != nil {
		t.Error(err)
	}
}

func TestAccountNewNotHuman(t *testing.T) {
	userModel := newMockUserModel()
	accountEndpoints := &AccountEndpoints{
		userModel, &mockSender{}, &mockChecker{},
	}

	buf := bytes.NewBuffer([]byte(`{
		"username": "user",
		"email": "email@example.com",
		"password": "password",
		"captcha": "cheater"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	accountEndpoints.AccountNew(w, req)

	if w.Result().StatusCode != 403 {
		t.Error("New account did not return 403")
	}

	if err := userModel.ExpectationsWereMetInOrder(); err != nil {
		t.Error(err)
	}
}

func TestAccountNewBadInputs(t *testing.T) {
	userModel := newMockUserModel()
	accountEndpoints := &AccountEndpoints{
		userModel, &mockSender{}, &mockChecker{},
	}

	testCases := []string{
		`{
			"username": "",
			"email": "email@example.com",
			"password": "password",
			"captcha": "12345"
		}`,
		`{
			"username": "user",
			"email": "",
			"password": "password",
			"captcha": "12345"
		}`,
		`{
			"username": "user",
			"email": "email@example.com",
			"password": "",
			"captcha": "12345"
		}`,
		`{
			"username": "012345678901234567890",
			"email": "email@example.com",
			"password": "",
			"captcha": "12345"
		}`,
		`{
			"username": "user",
			"email": "email@example.com",
			"password": "short",
			"captcha": "12345"
		}`,
	}

	for _, test := range testCases {
		buf := bytes.NewBuffer([]byte(test))
		req := httptest.NewRequest("POST", "http://localhost", buf)
		w := httptest.NewRecorder()
		accountEndpoints.AccountNew(w, req)

		if w.Result().StatusCode != 400 {
			t.Error("New account did not return 400")
		}

		if err := userModel.ExpectationsWereMetInOrder(); err != nil {
			t.Error(err)
		}
	}
}

func TestAccountNewUsernameExists(t *testing.T) {
	userModel := newMockUserModel()
	accountEndpoints := &AccountEndpoints{
		userModel, &mockSender{}, &mockChecker{},
	}

	userModel.ExpectSelectByUsername().WithArgs("user").WillReturn(testUser())

	buf := bytes.NewBuffer([]byte(`{
		"username": "user",
		"email": "email@example.com",
		"password": "password",
		"captcha": "12345"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	accountEndpoints.AccountNew(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("New account did not return 400")
	}

	if err := userModel.ExpectationsWereMetInOrder(); err != nil {
		t.Error(err)
	}
}

func TestAccountNewEmailExists(t *testing.T) {
	userModel := newMockUserModel()
	accountEndpoints := &AccountEndpoints{
		userModel, &mockSender{}, &mockChecker{},
	}

	userModel.ExpectSelectByUsername().WithArgs("user")
	userModel.ExpectSelectByEmail().WithArgs("email@example.com").
		WillReturn(testUser())

	buf := bytes.NewBuffer([]byte(`{
		"username": "user",
		"email": "email@example.com",
		"password": "password",
		"captcha": "12345"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	accountEndpoints.AccountNew(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("New account did not return 400")
	}

	if err := userModel.ExpectationsWereMetInOrder(); err != nil {
		t.Error(err)
	}
}

func TestAccountExistsUsername(t *testing.T) {

	testCases := map[string]string{
		"name":      "true",
		"invisible": "false",
	}

	userModel := newMockUserModel()
	r := mux.NewRouter()
	accountEndpoints := NewAccountEndpoints(userModel, nil, nil)
	accountEndpoints.Attach(r)
	for username, expected := range testCases {
		url := "http://localhost/account/exists/username/" + username
		req := httptest.NewRequest("GET", url, nil)

		if username == "name" {
			userModel.ExpectSelectByUsername().WithArgs("name").WillReturn(testUser())
		} else {
			userModel.ExpectSelectByUsername().WithArgs(username)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Error(err)
		}

		if string(body) != expected {
			t.Error("Test account exists by username did not return the right value!")
		}

		if err := userModel.ExpectationsWereMetInOrder(); err != nil {
			t.Error(err)
		}
	}
}

func TestAccountExistsEmail(t *testing.T) {

	testCases := map[string]string{
		"email@example.com": "true",
		"blah@example.com":  "false",
	}

	userModel := newMockUserModel()
	r := mux.NewRouter()
	accountEndpoints := NewAccountEndpoints(userModel, nil, nil)
	accountEndpoints.Attach(r)
	for email, expected := range testCases {
		url := "http://localhost/account/exists/email/" + email
		req := httptest.NewRequest("GET", url, nil)

		if email == "email@example.com" {
			fmt.Println(email)
			userModel.ExpectSelectByEmail().WithArgs(email).WillReturn(testUser())
		} else {
			userModel.ExpectSelectByEmail().WithArgs(email)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Error(err)
		}

		if string(body) != expected {
			t.Error("Test account exists by email did not return " + expected)
		}

		if err := userModel.ExpectationsWereMetInOrder(); err != nil {
			t.Error(err)
		}
	}
}
