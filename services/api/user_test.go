package api

import (
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
	"time"
)

func TestGetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	columns := []string{"id", "username", "email", "bio", "joined",
		"permission_set", "verified"}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE username = (.+)`,
	).
		WithArgs("us3rn4m3").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "us3rn4m3", "email@example.com", "Hello, World!", time.Now(),
			"USER", true,
		))

	user, err := GetUser(db, "us3rn4m3")
	if err != nil {
		t.Error(err)
	}

	if user.Id != 1 {
		t.Error("User id was wrong!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetNilUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	columns := []string{"id", "username", "email", "bio", "joined",
		"permission_set", "verified"}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE username = (.+)`,
	).
		WithArgs("us3rn4m3").
		WillReturnRows(sqlmock.NewRows(columns))

	user, err := GetUser(db, "us3rn4m3")
	if err != nil {
		t.Error(err)
	}

	if user != nil {
		t.Error("User should be nil!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAddUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE username = (.+)`,
	).
		WithArgs("username").
		WillReturnRows(sqlmock.NewRows([]string{"num"}))

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE email = (.+)`,
	).
		WithArgs("email@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"num"}))

	mock.ExpectExec(
		"WITH u AS \\(( INSERT INTO users (.+) VALUES (.+) RETURNING id )\\) "+
			"INSERT INTO passwords (.+) VALUES (.+)",
	).WithArgs(
		"username", "email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = AddUser(db, "username", "password", "email@example.com")

	if err != nil {
		t.Error(err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAddUserUserExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE username = (.+)`,
	).
		WithArgs("username").
		WillReturnRows(sqlmock.NewRows([]string{"num"}).AddRow(1))

	err = AddUser(db, "username", "password", "email@example.com")

	if err != UserExistsError {
		t.Error("AddUser did not reject existing user!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAddUserEmailExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE username = (.+)`,
	).
		WithArgs("username").
		WillReturnRows(sqlmock.NewRows([]string{"num"}))

	mock.ExpectQuery(
		`SELECT 1 FROM users WHERE email = (.+)`,
	).
		WithArgs("email@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"num"}).AddRow(1))

	err = AddUser(db, "username", "password", "email@example.com")

	if err != EmailExistsError {
		t.Error("AddUser did not reject existing email!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAddUserShortPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	err = AddUser(db, "username", "passw", "email@example.com")

	if err != PasswordToShortError {
		t.Error("AddUser did not reject short password!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
