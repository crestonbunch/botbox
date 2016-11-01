package api

import (
	"database/sql"
	"net/http"
	"time"
)

// A session type stores information about the current user making an API
// request. It allows handlers to retrieve preprocessed information easily about
// the current user or session without needing to do extra work.
type Session struct {
	Secret    string
	User      *User
	UserAgent string
	Ip        string
	Created   time.Time
	Expires   time.Time
	Revoked   bool
}

// Get a current (valid) session from the database, or return nil if no session
// was found.
// Returns a joined user struct within the session, so the current user can
// be retrieved without an extra query.
func GetSession(db *sql.DB, secret string) (*Session, error) {
	u := User{}
	s := Session{User: &u}
	err := db.QueryRow(
		`SELECT
			s.secret, s.user_agent, s.ip, s.created, s.expires, s.revoked,
			u.id, u.username, u.email, u.bio, u.joined
			u.permissions
		FROM
			user_sessions AS s, users as u
		WHERE
			secret = ?
			AND s.revoked != FALSE
			AND NOW() < s.expires
			AND u.id = s.user
		`,
		secret,
	).Scan(
		&s.Secret,
		&s.UserAgent,
		&s.Ip,
		&s.Created,
		&s.Expires,
		&s.Revoked,
		&u.Id,
		&u.Username,
		&u.Email,
		&u.Bio,
		&u.Joined,
		&u.Verified,
		&u.PermissionSet,
	)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &s, nil
	}
}

func SessionAuth(w http.ResponseWriter, r *http.Request) {

}

func SessionRenew(w http.ResponseWriter, r *http.Request) {

}

func SessionRevoke(w http.ResponseWriter, r *http.Request) {

}

func SessionUser(w http.ResponseWriter, r *http.Request) {

}
