package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// A basic handler that will stuff some Json into an interface.
type JsonHandler struct {
	// The target interface to fit JSON into. Pass in a factory function
	// which creates a new empty struct pointer to fill with values.
	Target func() interface{}
}

func (h *JsonHandler) Handle(r *http.Request) (interface{}, *HttpError) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	m := h.Target()
	err = json.Unmarshal(body, m)
	if err != nil {
		log.Println(err)
		return nil, ErrInvalidJson
	}

	return m, nil
}

// A handler that also authenticates the user from the HTTP headers.
type JsonHandlerWithAuth struct {
	session     SessionModel
	token       string
	Target      func() interface{}
	User        int
	Permissions PermissionSet
}

// Extract the user Id and insert into the User property
func (h *JsonHandlerWithAuth) HandleWithId(r *http.Request) (interface{}, *HttpError) {
	parent := &JsonHandler{Target: h.Target}
	target, parseerr := parent.Handle(r)
	if parseerr != nil {
		return nil, parseerr
	}

	h.token = ParseAuthSecret(r.Header)
	user, err := h.session.GetUserId(h.token)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}
	h.User = user

	return target, nil
}

// Extract the user id and permissions
func (h *JsonHandlerWithAuth) HandleWithPermissions(r *http.Request) (interface{}, *HttpError) {
	parent := &JsonHandler{Target: h.Target}
	target, parseerr := parent.Handle(r)
	if parseerr != nil {
		return nil, parseerr
	}

	h.token = ParseAuthSecret(r.Header)
	user, err := h.session.GetUserId(h.token)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}
	h.User = user

	permissions, err := h.session.GetPermissions(h.token)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}
	h.Permissions = permissions

	return target, nil
}

// URLPathHandler takes URL parameters of the form /path/{var1}/{var2} extracted
// by a gorilla mux and converts them into JSON then unmarshals them into the
// target interface
type URLPathHandler struct {
	// The target interface to fit URL parameters into. Pass in a factory function
	// which creates a new empty struct pointer to fill with values.
	Target func() interface{}
}

// Handle takes the path and puts the URL parameters into the target interfaced
// given to the URLPathHandler struct.
func (h *URLPathHandler) Handle(r *http.Request) (interface{}, *HttpError) {
	vars := mux.Vars(r)

	body, err := json.Marshal(vars)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	m := h.Target()
	err = json.Unmarshal(body, m)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	return m, nil
}
