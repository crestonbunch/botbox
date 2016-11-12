package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"io"
	"strings"
	"time"
)

var SessionColumns = []string{
	"secret", "user", "created", "expires", "revoked",
}

const (
	SessionSecretLen = 64
)

type Session struct {
	Secret  string
	User    int
	Created time.Time
	Expires time.Time
	Revoked bool
}

type SessionModel interface {

	// Get a session with the given secret
	GetSession(secret string) (*Session, error)

	// Start a new session for a user. Note that this does not perform any
	// validations. Those should be done before by the user model.
	StartSession(username string) (*Session, error)

	// Start a new session for the user that expires at the given time.
	StartSessionWithExpiration(username string, exp time.Time) (*Session, error)

	// Generate a new session for one that is expiring soon.
	RenewSession(secret string) (*Session, error)

	// Revoke a session
	RevokeSession(secret string) (*Session, error)
}

// This is the default sessions model, which handles session management through
// the database.
type Sessions struct {
	db *sql.DB
}

// Get a session with the secret provided. Returns an error if no such session
// exists, or the session is expired or revoked.
func (s *Sessions) GetSession(secret string) (*Session, error) {
	sess := Session{}
	err := s.db.QueryRow(
		`SELECT `+
			strings.Join(SessionColumns, ",")+`
		FROM
			user_sessions
		WHERE
			secret = $1
			AND expires < NOW()
			AND revoked = false
		`,
		secret,
	).Scan(
		&sess.Secret,
		&sess.User,
		&sess.Created,
		&sess.Expires,
		&sess.Revoked,
	)

	if err != nil {
		return nil, err
	} else {
		return &sess, nil
	}
}

func (s *Sessions) StartSession(username string) (*Session, error) {
	raw := make([]byte, SessionSecretLen)
	_, err := io.ReadFull(rand.Reader, raw)
	if err != nil {
		return nil, err
	}

	secret := make([]byte, base64.RawURLEncoding.EncodedLen(len(raw)))
	base64.RawURLEncoding.Encode(secret, raw)

	_, err = s.db.Exec(
		`INSERT INTO user_sessions (secret, user) VALUES
		$1, (SELECT id FROM users WHERE username = $2)`,
		secret,
		username,
	)

	return s.GetSession(string(secret))
}
