package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHmacValidate(t *testing.T) {
	buf := bytes.NewBuffer([]byte("Hello, World!"))
	date := time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST")
	secret := "r4nd0mS3CR3t"
	req := httptest.NewRequest("POST", "http://localhost", buf)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(date + secret + buf.String()))
	message := string(mac.Sum(nil))
	req.Header.Set("Date", date)
	req.Header.Set("X-Request-ID", secret)
	req.Header.Set("Authorization", "HMAC-SHA256 "+message)

	if !hmacValidate(secret, message, req) {
		t.Error("HMAC did not validate correctly!")
	}
}

func TestHmacTimeout(t *testing.T) {
	buf := bytes.NewBuffer([]byte("Hello, World!"))
	date := time.Now().Add(-20 * time.Second).
		Format("Mon, 02 Jan 2006 15:04:05 MST")
	secret := "r4nd0mS3CR3t"
	req := httptest.NewRequest("POST", "http://localhost", buf)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(date + secret + buf.String()))
	message := string(mac.Sum(nil))
	req.Header.Set("Date", date)
	req.Header.Set("X-Request-ID", secret)
	req.Header.Set("Authorization", "HMAC-SHA256 "+message)

	if hmacValidate(secret, message, req) {
		t.Error("HMAC should not validate correctly!")
	}
}

func TestHmacBadSecret(t *testing.T) {
	buf := bytes.NewBuffer([]byte("Hello, World!"))
	date := time.Now().Add(-20 * time.Second).
		Format("Mon, 02 Jan 2006 15:04:05 MST")
	secret := "r4nd0mS3CR3t"
	req := httptest.NewRequest("POST", "http://localhost", buf)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(date + secret + buf.String()))
	message := string(mac.Sum(nil))
	req.Header.Set("Date", date)
	req.Header.Set("X-Request-ID", secret)
	req.Header.Set("Authorization", "HMAC-SHA256 "+message)

	if hmacValidate("wrongSecret", message, req) {
		t.Error("HMAC should not validate correctly!")
	}
}

func TestHmacBadMessage(t *testing.T) {
	buf := bytes.NewBuffer([]byte("Hello, World!"))
	date := time.Now().Add(-20 * time.Second).
		Format("Mon, 02 Jan 2006 15:04:05 MST")
	secret := "r4nd0mS3CR3t"
	req := httptest.NewRequest("POST", "http://localhost", buf)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(date + secret + "a" + buf.String()))
	message := string(mac.Sum(nil))
	req.Header.Set("Date", date)
	req.Header.Set("X-Request-ID", secret)
	req.Header.Set("Authorization", "HMAC-SHA256 "+message)

	if hmacValidate("wrongSecret", message, req) {
		t.Error("HMAC should not validate correctly!")
	}
}

func TestGetSessionFromHmac(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer([]byte("Hello, World!"))
	date := time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST")
	secret := "r4nd0mS3CR3t"
	req := httptest.NewRequest("POST", "http://localhost", buf)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(date + secret + buf.String()))
	message := string(mac.Sum(nil))
	req.Header.Set("Date", date)
	req.Header.Set("X-Request-ID", secret)
	req.Header.Set("Authorization", "HMAC-SHA256 "+message)

	columns := []string{"s.secret", "s.user_agent", "s.ip", "s.created",
		"s.expires", "s.revoked", "u.id", "u.username", "u.email",
		"u.bio", "u.joined", "u.verified", "u.permissions"}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE secret = (.+)`,
	).
		WithArgs(secret).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			secret, "browser", "127.0.0.1", time.Now(),
			time.Now().Add(10*time.Second), false, 1, "username",
			"email@example.com", "Hello, World!", time.Now(), true, "USER",
		))

	sess, err := getSession(db, req)
	if err != nil {
		t.Error(err)
	}
	if sess == nil {
		t.Error("Session is nil!")
	}

	if sess.Secret != secret {
		t.Error("Session secret was wrong!")
	}

	if sess.User.Id != 1 {
		t.Error("Session user was wrong!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMiddleWare(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer([]byte("Hello, World!"))
	date := time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST")
	secret := "r4nd0mS3CR3t"
	req := httptest.NewRequest("POST", "http://localhost", buf)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(date + secret + buf.String()))
	message := string(mac.Sum(nil))
	req.Header.Set("Date", date)
	req.Header.Set("X-Request-ID", secret)
	req.Header.Set("Authorization", "HMAC-SHA256 "+message)

	columns := []string{"s.secret", "s.user_agent", "s.ip", "s.created",
		"s.expires", "s.revoked", "u.id", "u.username", "u.email",
		"u.bio", "u.joined", "u.verified", "u.permissions"}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) AS s WHERE secret = (.+) INNER JOIN (.+)`,
	).
		WithArgs(secret).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			secret, "browser", "127.0.0.1", time.Now(),
			time.Now().Add(10*time.Second), false, 1, "username",
			"email@example.com", "Hello, World!", time.Now(), true, "USER",
		))

	handler := RequestWrapper(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			db := GetRequestDb(r)
			sess := GetRequestSession(r)
			user := GetRequestUser(r)

			if db == nil {
				t.Error("Database was nil in middleware!")
			}
			if sess.Secret != secret {
				t.Error("Secret was not set by middleware!")
			}
			if user.Id != 1 {
				t.Error("User was not set by middleware!")
			}
		}), db)

	w := httptest.NewRecorder()
	handler(w, req)
}
