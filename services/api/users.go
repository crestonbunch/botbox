package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"time"
)

const SaltLength = 32

var UserExistsError = errors.New("Username already exists!")
var EmailExistsError = errors.New("That email is already in use!")
var PasswordToShortError = errors.New("The password is too short!")

type User struct {
	Id            int
	Username      string
	Email         string
	Bio           string
	Joined        time.Time
	PermissionSet string
	Verified      bool
}

// Given a password, generate a salt and hash it and return the salt and
// hash and any error encountered.
func hashPassword(passw string) ([]byte, []byte, error) {
	raw := make([]byte, SaltLength)
	_, err := io.ReadFull(rand.Reader, raw)
	if err != nil {
		return nil, nil, err
	}

	salt := make([]byte, base64.RawURLEncoding.EncodedLen(len(raw)))
	base64.RawURLEncoding.Encode(salt, raw)

	combined := append(salt, passw...)
	hash, err := bcrypt.GenerateFromPassword(combined, bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	return salt, hash, nil
}

// Compare a plaintext password with one that has been salted and hashed with
// the given salt and hash. Returns true if the passwords match.
func comparePassword(salt, hash, passw string) (bool, error) {
	combined := append([]byte(salt), []byte(passw)...)
	err := bcrypt.CompareHashAndPassword([]byte(hash), combined)
	if err == nil {
		return true, nil
	} else if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	} else {
		return false, err
	}
}

// Get a user by username. Returns nil if no user was found.
func GetUser(db *sql.DB, username string) (*User, error) {
	u := User{}
	err := db.QueryRow(
		`SELECT
			id, username, email, bio, joined, permission_set, verified
		FROM
			users
		WHERE
			username = $1
		`,
		username,
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
}

// Create a new unvalidated user. Simply adds a database entry. It does not
// send a confirmation email.
func AddUser(db *sql.DB, username, password, email string) error {
	salt, hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	if len(password) < MinPasswordLen {
		return PasswordToShortError
	}

	num := 0

	err = db.QueryRow(
		`SELECT 1 FROM users WHERE username = $1`, username,
	).Scan(&num)
	if err == nil {
		return UserExistsError
	}

	err = db.QueryRow(
		`SELECT 1 FROM users WHERE email = $1`, email,
	).Scan(&num)
	if err == nil {
		return EmailExistsError
	}

	_, err = db.Exec(
		`WITH u AS (
			INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id
		)
		INSERT INTO passwords ("user", method, hash, salt)
		VALUES ((SELECT id FROM u), 'bcrypt', $3, $4)`,
		username,
		email,
		string(hash),
		string(salt),
	)

	return err
}

func UserGet(w http.ResponseWriter, r *http.Request) {

}
