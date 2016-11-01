package api

import (
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
	"time"
)

func TestGetSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	columns := []string{"s.secret", "s.user_agent", "s.ip", "s.created",
		"s.expires", "s.revoked", "u.id", "u.username", "u.email",
		"u.bio", "u.joined", "u.verified", "u.permissions"}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE secret = (.+)`,
	).
		WithArgs("s3cr3t").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			"s3cr3t", "browser", "127.0.0.1", time.Now(),
			time.Now().Add(10*time.Second), false, 1, "username",
			"email@example.com", "Hello, World!", time.Now(), true, "USER",
		))

	sess, err := GetSession(db, "s3cr3t")
	if err != nil {
		t.Error(err)
	}

	if sess.Secret != "s3cr3t" {
		t.Error("Session secret was wrong!")
	}

	if sess.User.Id != 1 {
		t.Error("Session user was wrong!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetNilSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	columns := []string{"s.secret", "s.user_agent", "s.ip", "s.created",
		"s.expires", "s.revoked", "u.id", "u.username", "u.email",
		"u.bio", "u.joined", "u.verified", "u.permissions"}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE secret = (.+)`,
	).
		WithArgs("s3cr3t").
		WillReturnRows(sqlmock.NewRows(columns))

	sess, err := GetSession(db, "s3cr3t")
	if err != nil {
		t.Error(err)
	}

	if sess != nil {
		t.Error("Session should be nil!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
