package api

import (
	"bytes"
	"database/sql"
	"errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"net/http/httptest"
	"testing"
)

func TestAccountNew(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	columns := []string{"id", "username", "email", "bio", "joined",
		"permission_set", "verified"}

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE username = (.+)`,
	).
		WithArgs("username").
		WillReturnRows(sqlmock.NewRows(columns))

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE email = (.+)`,
	).
		WithArgs("email@example.com").
		WillReturnRows(sqlmock.NewRows(columns))

	mock.ExpectExec(
		"WITH u AS \\(( INSERT INTO users (.+) VALUES (.+) RETURNING id )\\) "+
			"INSERT INTO passwords (.+) VALUES (.+)",
	).WithArgs(
		"username", "email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	handler := RequestWrapper(
		AccountNew(func(db *sql.DB, username, email string) error {
			return nil
		}), db)

	buf := bytes.NewBuffer([]byte(`{
		"username": "username",
		"email": "email@example.com",
		"password": "sup3rs3cr3t"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 200 {
		t.Error("New account did not return 200")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountNewInvalidBody(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	handler := RequestWrapper(
		AccountNew(func(db *sql.DB, username, email string) error {
			return nil
		}), db)

	buf := bytes.NewBuffer([]byte(`{
		"username": "username",
		"email": "email@example.com",
		"pazzword": "sup3rs3cr3t"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("New account did not return 400")
	}

}

func TestAccountNewInvalidJson(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	handler := RequestWrapper(
		AccountNew(func(db *sql.DB, username, email string) error {
			return nil
		}), db)

	buf := bytes.NewBuffer([]byte(`{
		"username": "username",
		"email": "email@example.com",
		"password": "sup3rs3cr3t",
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("New account did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountNewDatabaseErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	columns := []string{"id", "username", "email", "bio", "joined",
		"permission_set", "verified"}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE username = (.+)`,
	).
		WithArgs("username").
		WillReturnRows(sqlmock.NewRows(columns))

	mock.ExpectExec(
		"WITH u AS \\(( INSERT INTO users (.+) VALUES (.+) RETURNING id )\\) "+
			"INSERT INTO passwords (.+) VALUES (.+)",
	).WithArgs(
		"username", "email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnError(errors.New("Dummy"))

	handler := RequestWrapper(
		AccountNew(func(db *sql.DB, username, email string) error {
			return nil
		}), db)

	buf := bytes.NewBuffer([]byte(`{
		"username": "username",
		"email": "email@example.com",
		"password": "sup3rs3cr3t"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 500 {
		t.Error("New account did not return 500")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}

}

func TestAccountNewGetUserDatabaseErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE username = (.+)`,
	).
		WithArgs("username").
		WillReturnError(errors.New("Dummy"))

	handler := RequestWrapper(
		AccountNew(func(db *sql.DB, username, email string) error {
			return nil
		}), db)

	buf := bytes.NewBuffer([]byte(`{
		"username": "username",
		"email": "email@example.com",
		"password": "sup3rs3cr3t"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 500 {
		t.Error("New account did not return 500")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}

}

func TestAccountNewUserExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE username = (.+)`,
	).
		WithArgs("username").
		WillReturnRows(sqlmock.NewRows([]string{"nums"}).AddRow(1))

	handler := RequestWrapper(
		AccountNew(func(db *sql.DB, username, email string) error {
			return nil
		}), db)

	buf := bytes.NewBuffer([]byte(`{
		"username": "username",
		"email": "email@example.com",
		"password": "sup3rs3cr3t"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("New account did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountEmailExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE username = (.+)`,
	).
		WithArgs("username").
		WillReturnRows(sqlmock.NewRows([]string{"nums"}))

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE email = (.+)`,
	).
		WithArgs("email@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"nums"}).AddRow(1))

	handler := RequestWrapper(
		AccountNew(func(db *sql.DB, username, email string) error {
			return nil
		}), db)

	buf := bytes.NewBuffer([]byte(`{
		"username": "username",
		"email": "email@example.com",
		"password": "sup3rs3cr3t"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("New account did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountShortPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	handler := RequestWrapper(
		AccountNew(func(db *sql.DB, username, email string) error {
			return nil
		}), db)

	buf := bytes.NewBuffer([]byte(`{
		"username": "username",
		"email": "email@example.com",
		"password": "12345"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("New account did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountVerify(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(
		"UPDATE users u SET permission_set = 'VERIFIED', v.used = TRUE " +
			"FROM user_verifications v " +
			"WHERE " +
			"v.user = u.id AND v.secret = (.+) AND v.expires < now\\(\\) " +
			"AND v.used = FALSE AND v.permission_set = 'UNVERIFIED",
	).WithArgs(
		"s3cr3t",
	).WillReturnResult(sqlmock.NewResult(1, 1))

	handler := RequestWrapper(AccountVerify, db)

	buf := bytes.NewBuffer([]byte(`{
		"secret": "s3cr3t"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 200 {
		t.Error("Verify account did not return 200")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountVerifyErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(
		"UPDATE users u SET permission_set = 'VERIFIED', v.used = TRUE " +
			"FROM user_verifications v " +
			"WHERE " +
			"v.user = u.id AND v.secret = (.+) AND v.expires < now\\(\\) " +
			"AND v.used = FALSE AND v.permission_set = 'UNVERIFIED",
	).WithArgs(
		"s3cr3t",
	).WillReturnError(errors.New("Dummy"))

	handler := RequestWrapper(AccountVerify, db)

	buf := bytes.NewBuffer([]byte(`{
		"secret": "s3cr3t"
	}`))
	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Result().StatusCode != 400 {
		t.Error("Verify account did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountChangePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT salt, hash FROM passwords WHERE "user" = (.+)`,
	).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"salt", "hash"}).AddRow(
			"123", "$2a$10$tew2rT8mdPYH73GN/KsjgOBNS4NYvFiNY1jQFib56ROvzrgF/Zdxi",
		))

	mock.ExpectExec(
		"UPDATE passwords SET salt = (.+), hash = (.+), updated = (.+)"+
			"WHERE \"user\" = (.+)",
	).WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	handler := AccountChangePassword

	user := &User{Id: 1, Username: "user123"}
	sess := &Session{User: user}

	buf := bytes.NewBuffer([]byte(`{
		"old": "password",
		"new": "p455w0rd"
	}`))

	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req.WithContext(upgradeContext(req.Context(), db, sess, user)))

	if w.Result().StatusCode != 200 {
		t.Error("Change password did not return 200")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountChangePasswordWrongPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT salt, hash FROM passwords WHERE "user" = (.+)`,
	).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"salt", "hash"}).AddRow(
			"123", "$2a$10$tew2rT8mdPYH73GN/KsjgOBNS4NYvFiNY1jQFib56ROvzrgF/Zdxi",
		))

	handler := AccountChangePassword

	user := &User{Id: 1, Username: "user123"}
	sess := &Session{User: user}

	buf := bytes.NewBuffer([]byte(`{
		"old": "passwd",
		"new": "p455w0rd"
	}`))

	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req.WithContext(upgradeContext(req.Context(), db, sess, user)))

	if w.Result().StatusCode != 400 {
		t.Error("Change password did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountChangePasswordTooShort(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT salt, hash FROM passwords WHERE "user" = (.+)`,
	).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"salt", "hash"}).AddRow(
			"123", "$2a$10$tew2rT8mdPYH73GN/KsjgOBNS4NYvFiNY1jQFib56ROvzrgF/Zdxi",
		))

	handler := AccountChangePassword

	user := &User{Id: 1, Username: "user123"}
	sess := &Session{User: user}

	buf := bytes.NewBuffer([]byte(`{
		"old": "password",
		"new": "passw"
	}`))

	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req.WithContext(upgradeContext(req.Context(), db, sess, user)))

	if w.Result().StatusCode != 400 {
		t.Error("Change password did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountChangePasswordNilUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	handler := AccountChangePassword

	var user *User = nil
	var sess *Session = nil

	buf := bytes.NewBuffer([]byte(`{
		"old": "password",
		"new": "passw"
	}`))

	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req.WithContext(upgradeContext(req.Context(), db, sess, user)))

	if w.Result().StatusCode != 403 {
		t.Error("Change password did not return 403")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountChangePasswordNoUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT salt, hash FROM passwords WHERE "user" = (.+)`,
	).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"salt", "hash"}))

	handler := AccountChangePassword

	user := &User{Id: 1, Username: "user123"}
	sess := &Session{User: user}

	buf := bytes.NewBuffer([]byte(`{
		"old": "password",
		"new": "passw"
	}`))

	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req.WithContext(upgradeContext(req.Context(), db, sess, user)))

	if w.Result().StatusCode != 400 {
		t.Error("Change password did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAccountChangePasswordInvalidJson(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	handler := AccountChangePassword

	user := &User{Id: 1, Username: "user123"}
	sess := &Session{User: user}

	buf := bytes.NewBuffer([]byte(`{
		"old": "password",
		"new": "password",
	}`))

	req := httptest.NewRequest("POST", "http://localhost", buf)
	w := httptest.NewRecorder()
	handler(w, req.WithContext(upgradeContext(req.Context(), db, sess, user)))

	if w.Result().StatusCode != 400 {
		t.Error("Change password did not return 400")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCreateVerification(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(
		"INSERT INTO user_verifications (.+) VALUES (.+)",
	).WithArgs(
		sqlmock.AnyArg(), "test@example.com", "user123",
	).WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = CreateVerification(db, "user123", "test@example.com")
	if err != nil {
		t.Error(err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}

}
