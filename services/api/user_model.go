package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"io"
	"strings"
	"time"
)

var UserColumns = []string{
	"id", "username", "fullname", "email", "bio", "organization", "location",
	"joined", "permission_set",
}

const (
	SaltLength               = 4
	VerificationSecretLength = 64
)

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

type User struct {
	Id            int
	Username      string
	Fullname      string
	Email         string
	Bio           string
	Organization  string
	Location      string
	Joined        time.Time
	PermissionSet string
}

type UserModel interface {
	SelectByUsername(string) (*User, error)
	SelectByEmail(string) (*User, error)
	SelectByUsernameAndEmail(username, email string) (*User, error)
	SelectBySession(session string) (*User, error)
	Insert(username, email, password string) (int, error)
	Update(user *User, fullname, email, bio, organization, location string) error
	VerifyPassword(username, password string) (bool, error)
	ChangePassword(user *User, password string) error
	ChangePermissions(user *User, permissions string) error
	CreateVerificationSecret(id int, email string) (string, error)
}

func NewUserModel(db *sql.DB) *Users {
	return &Users{db}
}

// This is the default user model, which is backed by an sql database for
// selecting users. For example, Users.SelectByUsername("username") will return
// a user with the username "username".
type Users struct {
	db *sql.DB
}

// Scan a result from the database and insert it into a user object. Returns
// a nil user object if there is no row.
func (db *Users) scanRow(row *sql.Row) (*User, error) {
	u := User{}
	if err := row.Scan(
		&u.Id,
		&u.Username,
		&u.Fullname,
		&u.Email,
		&u.Bio,
		&u.Organization,
		&u.Location,
		&u.Joined,
		&u.PermissionSet,
	); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &u, nil
}

// Gets a user with the matching username, or returns nil if no user exists
// with that username. An error is returned for any SQL errors.
func (users *Users) SelectByUsername(username string) (*User, error) {
	u, err := users.scanRow(users.db.QueryRow(
		`SELECT `+
			strings.Join(UserColumns, ",")+`
		FROM
			users
		WHERE
			username = $1
		`,
		username,
	))

	if err != nil {
		return nil, err
	} else {
		return u, nil
	}
}

// Gets a user with the matching email, or returns nil if no user exists
// with that email. An error is returned for any SQL errors.
func (users *Users) SelectByEmail(email string) (*User, error) {
	u, err := users.scanRow(users.db.QueryRow(
		`SELECT `+
			strings.Join(UserColumns, ",")+`
		FROM
			users
		WHERE
			email = $1
		`,
		email,
	))

	if err != nil {
		return nil, err
	} else {
		return u, nil
	}
}

// Gets a user with the given username and email, or returns nil if no user
// exists with that username/email combination. An error is returned for any
// sql errors.
func (users *Users) SelectByUsernameAndEmail(
	username string, email string,
) (*User, error) {
	u, err := users.scanRow(users.db.QueryRow(
		`SELECT `+
			strings.Join(UserColumns, ",")+`
		FROM
			users
		WHERE
			username = $1 AND
			email = $1
		`,
		username,
		email,
	))

	if err != nil {
		return nil, err
	} else {
		return u, nil
	}
}

// Gets the user associated with a specific session secret, or nill if no user
// exists with that session.
func (users *Users) SelectBySession(session string) (*User, error) {
	u, err := users.scanRow(users.db.QueryRow(
		`SELECT `+
			strings.Join(UserColumns, ",")+`
		FROM
			users, sessions
		WHERE
			users.id = sessions.user AND
			sessions.secret = $1
		`,
		session,
	))

	if err != nil {
		return nil, err
	} else {
		return u, nil
	}
}

// Insert the user into the database with the given password. Returns the
// user id inserted or an error.
func (users *Users) Insert(username, email, password string) (int, error) {

	salt, hash, err := hashPassword(password)
	if err != nil {
		return 0, err
	}

	_, err = users.db.Exec(
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

	if err != nil {
		return 0, err
	}

	user, err := users.SelectByUsername(username)

	if err != nil {
		return 0, err
	}

	return user.Id, nil
}

// Update the fullname, email, bio, organization, and location of a user.
func (users *Users) Update(
	user *User, fullname, email, bio, organization, location string,
) error {

	_, err := users.db.Exec(
		`UPDATE users SET
			fullname = $1, email = $2, bio = $3, organization = $4, location = $5
		WHERE
			id = $6`,
		fullname,
		email,
		bio,
		organization,
		location,
		user.Id,
	)

	return err
}

// Verify that a given password matches the password for the username.
func (users *Users) VerifyPassword(username, password string) (bool, error) {

	result := struct {
		Method string
		Salt   string
		Hash   string
	}{}

	err := users.db.QueryRow(
		`SELECT
			method, salt, hash
		FROM
			passwords, users
		WHERE
			users.id = passwords.user AND
			users.username = $1
		`,
		username,
	).Scan(
		&result.Method,
		&result.Salt,
		&result.Hash,
	)

	if err != nil {
		return false, err
	}

	if result.Method == "bcrypt" {
		given := []byte(result.Salt + password)
		expected := []byte(result.Hash)

		err := bcrypt.CompareHashAndPassword(expected, given)

		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		} else if err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		return false, nil
	}

}

// Change a user's password. Note that no password checks are performed in this
// method, and passwords should always be validated before calling this method.
func (users *Users) ChangePassword(user *User, password string) error {

	salt, hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	_, err = users.db.Exec(
		`UPDATE passwords SET
			method = 'bcrypt', salt = $1, hash = $2,
		WHERE
			user = $3`,
		salt,
		hash,
		user.Id,
	)

	return err
}

// Change the permission set of a user.
func (users *Users) ChangePermissions(user *User, permissions string) error {
	_, err := users.db.Exec(
		`UPDATE users SET
			permission_set = $1
		WHERE
			id = $2`,
		permissions,
		user.Id,
	)

	return err
}

// Create a verification secret in the database for users who need to verify
// their account.
func (users *Users) CreateVerificationSecret(id int, email string) (string, error) {
	secret := make([]byte, VerificationSecretLength)
	_, err := io.ReadFull(rand.Reader, secret)
	if err != nil {
		return "", err
	}
	secretEnc := base64.RawURLEncoding.EncodeToString(secret)
	_, err = users.db.Exec(
		`INSERT INTO user_verifications (secret, email, "user")
		VALUES ($1, $2, $3)`,
		secretEnc,
		email,
		id,
	)

	return secretEnc, err
}
