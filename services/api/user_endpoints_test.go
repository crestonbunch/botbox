package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestUserGet(t *testing.T) {

	testCases := map[*http.Request]User{
		httptest.NewRequest("GET", "http://localhost/user/name", nil): *testUser(),
	}

	for req, expected := range testCases {
		model := newMockUserModel()
		userEndpoints := &UserEndpoints{model}

		r := mux.NewRouter()
		r.HandleFunc("/user/{username}", userEndpoints.UserGet).Methods("GET")

		model.ExpectSelectByUsername().WithArgs("name").WillReturn(testUser())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Error(err)
		}

		user := User{}
		err = json.Unmarshal(body, &user)
		if err != nil {
			t.Error(err)
		}

		if w.Result().StatusCode != 200 {
			t.Error("User get did not return 200")
		} else if !reflect.DeepEqual(user, expected) {
			t.Error("Get user returned the wrong user!")
		}

		if err := model.ExpectationsWereMetInOrder(); err != nil {
			t.Error(err)
		}
	}
}

func TestUserGetNoUser(t *testing.T) {

	req := httptest.NewRequest("GET", "http://localhost/user/name", nil)

	model := newMockUserModel()
	userEndpoints := &UserEndpoints{model}

	r := mux.NewRouter()
	r.HandleFunc("/user/{username}", userEndpoints.UserGet).Methods("GET")

	model.ExpectSelectByUsername().WithArgs("name")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("User get did not return 400")
	}

	if err := model.ExpectationsWereMetInOrder(); err != nil {
		t.Error(err)
	}
}
