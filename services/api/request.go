package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	ApiContextDB      = "database"
	ApiContextSession = "session"
	ApiContextUser    = "user"
	TimeoutSecs       = 10 //seconds
)

func hmacValidate(secret, expected string, req *http.Request) bool {
	dateStr := req.Header.Get("Date")
	if dateStr == "" {
		return false
	}
	date, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", dateStr)
	if err != nil {
		return false
	}
	if time.Since(date) > TimeoutSecs*time.Second {
		return false
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return false
	}

	// concatenate parts of the message together to hash with the secret and
	// verify the integrity of the message
	message := []byte(dateStr + secret + string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(message)
	return hmac.Equal(mac.Sum(nil), []byte(expected))
}

// Validates the session ID from the client and gets the current session from
// the database.
func getSession(db *sql.DB, req *http.Request) (*Session, error) {
	auth := req.Header.Get("Authorization")
	secret := req.Header.Get("X-Request-ID")
	if auth == "" || secret == "" {
		return nil, nil
	}

	if strings.HasPrefix(auth, "HMAC-SHA256 ") {
		authStr := strings.TrimPrefix(auth, "HMAC-SHA256 ")
		if !hmacValidate(secret, authStr, req) {
			return nil, nil
		}
		return GetSession(db, secret)
	}

	return nil, nil
}

// Upgrades a context to include values for the database, current user session,
// and user information.
func upgradeContext(
	ctx context.Context, db *sql.DB, sess *Session, user *User,
) context.Context {
	withDb := context.WithValue(ctx, ApiContextDB, db)
	withSession := context.WithValue(withDb, ApiContextSession, sess)
	withUser := context.WithValue(withSession, ApiContextUser, user)
	return withUser
}

// Wrap HTTP handlers with this middleware in order to give them information
// about the current session information from the database.
func RequestWrapper(
	next func(http.ResponseWriter, *http.Request), db *sql.DB,
) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sess, err := getSession(db, req)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if sess == nil {
			sess = &Session{}
		}
		user := sess.User
		ctx := upgradeContext(req.Context(), db, sess, user)
		next(w, req.WithContext(ctx))
	})
}

func GetRequestDb(req *http.Request) *sql.DB {
	return req.Context().Value(ApiContextDB).(*sql.DB)
}

func GetRequestSession(req *http.Request) *Session {
	return req.Context().Value(ApiContextSession).(*Session)
}

func GetRequestUser(req *http.Request) *User {
	return req.Context().Value(ApiContextUser).(*User)
}
