package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	MinPasswordLen           = 6
	VerificationSecretLength = 64
)

// Insert a record in the database for a user to verify his/her email address.
func CreateVerification(db *sql.DB, username, email string) (string, error) {
	secret := make([]byte, VerificationSecretLength)
	_, err := io.ReadFull(rand.Reader, secret)
	if err != nil {
		return "", err
	}
	secretEnc := base64.RawURLEncoding.EncodeToString(secret)
	_, err = db.Exec(
		`INSERT INTO user_verifications (secret, email, "user")
		VALUES ($1, $2, (SELECT id FROM users WHERE username = $3))`,
		secretEnc,
		email,
		username,
	)

	return secretEnc, err
}

// Create a new account in the database. This function wraps the actual
// http handler function, requiring a callback which will request validation
// from the user via email (or automatically validate, etc.)
// {
//		"username": "john_doe",
//		"password": "p455w0rd$$",
//		"email":		"email@example.com"
// }
// Returns 400 errors for malformed requests or invalid username/passwords.
func AccountNew(
	sendValidation func(db *sql.DB, username, email string) error,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := GetRequestDb(r)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		msg := struct {
			Username string `json:'username'`
			Password string `json:'password'`
			Email    string `json:'email'`
		}{}

		err = json.Unmarshal(body, &msg)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid JSON", 400)
			return
		}
		if msg.Username == "" || msg.Password == "" || msg.Email == "" {
			http.Error(w, "No username, password, or email.", 400)
			return
		}

		err = AddUser(db, msg.Username, msg.Password, msg.Email)
		if err == UserExistsError {
			http.Error(w, "That username already exists!", 400)
			return
		} else if err == EmailExistsError {
			http.Error(w, "That email is already is use!", 400)
			return
		} else if err == PasswordToShortError {
			http.Error(w, fmt.Sprintf(
				"Password must be at least %d characters.", MinPasswordLen,
			), 400)
			return
		} else if err != nil {
			log.Println(err)
			http.Error(w, "An error occurred please try again.", 500)
			return
		}

		err = sendValidation(db, msg.Username, msg.Email)
		if err != nil {
			log.Println(err)
			http.Error(w, "An error occurred please try again.", 500)
			return
		}
	}
}

// Make a request to verify a user's account given a verification secret.
// Example request:
// {
//		secret: 111222333abcdef
// }
// If the secret matches a valid secret in the database, then the user
// corresponding to the secret will have permissions upgraded from
// UNVERIFIED to VERIFIED.
// Otherwise, if the secret does not exist / has been used / the user
// permissions are not UNVERIFIED, then a 400 error is returned.
func AccountVerify(w http.ResponseWriter, r *http.Request) {
	db := GetRequestDb(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	msg := struct {
		Secret string `json:"secret"`
	}{}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", 400)
		return
	}

	_, err = db.Exec(
		`UPDATE users u SET permission_set = 'VERIFIED', v.used = TRUE
		FROM user_verifications v
		WHERE
			v.user = u.id AND v.secret = $1 AND v.expires < now() AND v.used = FALSE
			AND v.permission_set = 'UNVERIFIED'`,
		msg.Secret,
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid verification secret.", 400)
		return
	}
}

// Make a request to change a user's password. The request must have sent
// the appropriate authentication secret in the headers.
// Example request:
// {
//		old: p4$$W0rD,
//		new: Pa55w0Rd2
// }
func AccountChangePassword(w http.ResponseWriter, r *http.Request) {
	db := GetRequestDb(r)
	user := GetRequestUser(r)

	if user == nil {
		http.Error(w, "Not authenticated.", 403)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Parse the request body
	msg := struct {
		Old string `json:"old"`
		New string `json:"new"`
	}{}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", 400)
		return
	}

	// verify old password with the one in the database
	pw := struct {
		Salt string
		Hash string
	}{}

	err = db.QueryRow(
		`SELECT salt, hash FROM passwords WHERE "user" = $1`, user.Id,
	).Scan(&pw.Salt, &pw.Hash)

	if err == sql.ErrNoRows {
		// This shouldn't happen, but if it does idk
		log.Printf("User password not found: %d\n", user.Id)
		http.Error(w, "User password not found.", 400)
		return
	} else if err != nil {
		http.Error(w, "An unknown error occurred. Please try again.", 500)
		return
	}

	if ok, err := comparePassword(pw.Salt, pw.Hash, msg.Old); err != nil {
		log.Println(err)
		http.Error(w, "An unknown error occurred. Please try again.", 500)
		return
	} else if !ok {
		http.Error(w, "Invalid password!", 400)
		return
	}

	// check if password meets length requirement
	if len(msg.New) < MinPasswordLen {
		http.Error(w, fmt.Sprintf(
			"Password must be at least %d characters.", MinPasswordLen,
		), 400)
		return
	}

	// update the new password
	salt, hash, err := hashPassword(msg.New)
	if err != nil {
		log.Println(err)
		http.Error(w, "An unknown error occurred. Please try again.", 500)
		return
	}

	_, err = db.Exec(
		`UPDATE passwords SET salt = $1, hash = $2, updated = $3 WHERE "user" = $4`,
		salt,
		hash,
		time.Now(),
		user.Id,
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to change password.", 500)
		return
	}
}

// Return an HTTP handler func that will create a forgot password recovery
// secret in the database. This is wrapped by a function which will provide
// a callback to send the recovery key to the user. A sample request looks like:
// {
//		"username": "user123"
//		"email": "email@example.com"
// }
// Note that username or email are both optional, but at least one must be
// provided.
func AccountForgotPassword(
	sendRecovery func(db *sql.DB, username, email string) error,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
			db := GetRequestDb(r)

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			// Parse the request body
			msg := struct {
				Username string `json:"username"`
				Email    string `json:"email"`
			}{}

			err = json.Unmarshal(body, &msg)
			if err != nil {
				log.Println(err)
				http.Error(w, "Invalid JSON", 400)
				return
			}

			u := User{}

			if msg.Username != "" && msg.Email != "" {
				// both username and email are provided
				err := db.QueryRow(
					`SELECT
						id, username, email, bio, joined, permission_set, verified
					FROM
						users
					WHERE
					username =$1
				`,
					msg.Username,
				).Scan(
					&u.Id,
					&u.Username,
					&u.Email,
					&u.Bio,
					&u.Joined,
					&u.PermissionSet,
					&u.Verified,
				)
				switch {
				case err == sql.ErrNoRows:
					return nil, nil
				case err != nil:
					return nil, err
				default:
					return &u, nil
				}

			} else if msg.Username != "" {
				// just email is provided
			} else if msg.Email != "" {
				// just email is provided
			} else {
				// neither are provided
				http.Error(w, "Please provide a username or password.", 400)
				return
			}
		*/

	}
}
