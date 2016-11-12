package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// UserEndpoints is a struct that holds methods for all endpoints that relate
// to users. To attach them to a router, you can call Attach().
type UserEndpoints struct {
	model UserModel
}

func NewUserEndpoints(userModel UserModel) *UserEndpoints {
	return &UserEndpoints{userModel}
}

// Attach the endpoints to a router.
func (e *UserEndpoints) Attach(r *mux.Router) {
	r.HandleFunc("/user/{username}", e.UserGet).Methods("GET")
}

// Get a user by username.
// e.g. /api/users/user1234
// example response:
// {
//   id: 1,
//   username: "user1234",
//   ...
// }
func (e *UserEndpoints) UserGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	user, err := e.model.SelectByUsername(username)

	if err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	} else if user == nil {
		http.Error(w, UserNotFound.Error(), 400)
		return
	}

	b, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	}

	fmt.Fprint(w, string(b))
}
