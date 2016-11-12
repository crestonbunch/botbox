package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

// AccountEndpoints is a struct that holds methods for all endpoints that relate
// to account management. To attach them to a router, you can call Attach().
type AccountEndpoints struct {
	userModel   UserModel
	emailSender EmailSender
	botChecker  BotChecker
}

func NewAccountEndpoints(
	userModel UserModel, emailSender EmailSender, botChecker BotChecker,
) *AccountEndpoints {
	return &AccountEndpoints{userModel, emailSender, botChecker}
}

// Attach the endpoints to a router.
func (e *AccountEndpoints) Attach(r *mux.Router) {
	r.HandleFunc("/account/new", e.AccountNew).Methods("POST")
	r.HandleFunc("/account/exists/username/{username}", e.AccountExistsUsername).
		Methods("GET")
	r.HandleFunc("/account/exists/email/{email}", e.AccountExistsEmail).
		Methods("GET")
}

// An interface to something that can verify the humanness of a user if
// given a string token from the form.
type BotChecker interface {
	IsHuman(string) (bool, error)
}

// An interface that can send emails to a given address.
type EmailSender interface {
	SendVerificationEmail(to string, secret string) error
	SendPasswordRecoveryEmail(to string, secret string) error
}

const (
	MinPasswordLen = 6
	MaxUsernameLen = 20
)

// Create a new account in the database.
// {
//		"username": "john_doe",
//		"password": "p455w0rd$$",
//		"email":		"email@example.com"
//		"captcha":	"recaptchS3cr3t"
// }
// Returns 400 errors for malformed requests or invalid username/passwords.
func (e *AccountEndpoints) AccountNew(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	msg := struct {
		Username string `json:'username'`
		Password string `json:'password'`
		Email    string `json:'email'`
		Captcha  string `json:'captcha'`
	}{}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		log.Println(err)
		http.Error(w, InvalidJson.Error(), 400)
		return
	}

	if ok, err := e.botChecker.IsHuman(msg.Captcha); err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	} else if !ok {
		log.Println("A bot has been detected!")
		http.Error(w, BotDetected.Error(), 403)
		return
	}

	if msg.Username == "" || msg.Password == "" || msg.Email == "" {
		http.Error(w, MissingNewAccountField.Error(), 400)
		return
	}

	if len(msg.Username) > MaxUsernameLen {
		http.Error(w, UsernameTooLong.Error(), 400)
	}

	if len(msg.Password) < MinPasswordLen {
		http.Error(w, PasswordTooShort.Error(), 400)
	}

	if user, err := e.userModel.SelectByUsername(msg.Username); err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	} else if user != nil {
		http.Error(w, UsernameExists.Error(), 400)
		return
	}

	if user, err := e.userModel.SelectByEmail(msg.Email); err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	} else if user != nil {
		http.Error(w, EmailInUse.Error(), 400)
		return
	}

	id, err := e.userModel.Insert(msg.Username, msg.Email, msg.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	}

	secret, err := e.userModel.CreateVerificationSecret(id, msg.Email)
	if err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	}

	err = e.emailSender.SendVerificationEmail(msg.Email, secret)
	if err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	}
}

// Checks if an account already exists with the given username.
func (e *AccountEndpoints) AccountExistsUsername(
	w http.ResponseWriter, r *http.Request,
) {

	vars := mux.Vars(r)
	username := vars["username"]

	user, err := e.userModel.SelectByUsername(username)

	if err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	} else if user == nil {
		fmt.Fprint(w, "false")
		return
	} else {
		fmt.Fprint(w, "true")
		return
	}
}

// Checks if an account already exists with the given email.
func (e *AccountEndpoints) AccountExistsEmail(
	w http.ResponseWriter, r *http.Request,
) {

	vars := mux.Vars(r)
	email := vars["email"]

	user, err := e.userModel.SelectByEmail(email)

	if err != nil {
		log.Println(err)
		http.Error(w, UnknownError.Error(), 500)
		return
	} else if user == nil {
		fmt.Fprint(w, "false")
		return
	} else {
		fmt.Fprint(w, "true")
		return
	}
}
