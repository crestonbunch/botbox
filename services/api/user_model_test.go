package api

import (
	"database/sql"
	"errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reflect"
	"testing"
	"time"
)

var t = time.Now()

func testRow() sqlmock.Rows {
	testRow := sqlmock.NewRows(UserColumns).
		AddRow(1, "name", "John Doe", "email@example.com", "Bio", "Botbox Corp.",
			"Narnia", t, "USER")

	return testRow
}

func testUser() *User {
	testUser := &User{
		1, "name", "John Doe", "email@example.com", "Bio", "Botbox Corp.",
		"Narnia", t, "USER",
	}

	return testUser
}

func TestUsersSelectByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	users := &Users{db}

	testCases := map[sqlmock.Rows]*User{
		// Test a valid user exists
		testRow(): testUser(),
		// Test user does not exist
		sqlmock.NewRows(UserColumns): nil,
	}

	for row, expected := range testCases {
		mock.ExpectQuery(
			`SELECT (.+) FROM (.+) WHERE username = (.+)`,
		).
			WithArgs("user").
			WillReturnRows(row)

		user, err := users.SelectByUsername("user")

		if err != nil {
			t.Error(err)
		}

		if expected == nil && user != nil {
			t.Error("Selected user is not nil!")
		} else if !reflect.DeepEqual(user, expected) {
			t.Error("Select by user returned the wrong user!")
		}
	}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE username = (.+)`,
	).
		WithArgs("user").
		WillReturnError(errors.New("Dummy error."))

	user, err := users.SelectByUsername("user")

	if user != nil {
		t.Error("Select by username should return nil user on error!")
	}

	if err == nil {
		t.Error("Select by username should return non-nil error on error!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUsersSelectByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	users := &Users{db}

	testCases := map[sqlmock.Rows]*User{
		// Test a valid user exists
		testRow(): testUser(),
		// Test user does not exist
		sqlmock.NewRows(UserColumns): nil,
	}

	for row, expected := range testCases {
		mock.ExpectQuery(
			`SELECT (.+) FROM (.+) WHERE email = (.+)`,
		).
			WithArgs("email@example.com").
			WillReturnRows(row)

		user, err := users.SelectByEmail("email@example.com")

		if err != nil {
			t.Error(err)
		}

		if expected == nil && user != nil {
			t.Error("Selected user is not nil!")
		} else if !reflect.DeepEqual(user, expected) {
			t.Error("Select by email returned the wrong user!")
		}
	}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE email = (.+)`,
	).
		WithArgs("email@example.com").
		WillReturnError(errors.New("Dummy error."))

	user, err := users.SelectByEmail("email@example.com")

	if user != nil {
		t.Error("Select by email should return nil user on error!")
	}

	if err == nil {
		t.Error("Select by email should return non-nil error on error!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUsersSelectByUsernameAndEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	users := &Users{db}

	testCases := map[sqlmock.Rows]*User{
		// Test a valid user exists
		testRow(): testUser(),
		// Test user does not exist
		sqlmock.NewRows(UserColumns): nil,
	}

	for row, expected := range testCases {
		mock.ExpectQuery(
			`SELECT (.+) FROM (.+) WHERE username = (.+) AND email = (.+)`,
		).
			WithArgs("user", "email@example.com").
			WillReturnRows(row)

		user, err := users.SelectByUsernameAndEmail("user", "email@example.com")

		if err != nil {
			t.Error(err)
		}

		if expected == nil && user != nil {
			t.Error("Selected user is not nil!")
		} else if !reflect.DeepEqual(user, expected) {
			t.Error("Select by user returned the wrong user!")
		}
	}

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE username = (.+) AND email = (.+)`,
	).
		WithArgs("user", "email@example.com").
		WillReturnError(errors.New("Dummy error."))

	user, err := users.SelectByUsernameAndEmail("user", "email@example.com")

	if user != nil {
		t.Error("Select by email should return nil user on error!")
	}

	if err == nil {
		t.Error("Select by email should return non-nil error on error!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUsersSelectBySession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	users := &Users{db}

	testCases := map[sqlmock.Rows]*User{
		// Test a valid user exists
		testRow(): testUser(),
		// Test user does not exist
		sqlmock.NewRows(UserColumns): nil,
	}

	for row, expected := range testCases {
		mock.ExpectQuery(
			"SELECT (.+) FROM (.+) WHERE users.id = sessions.user" +
				" AND sessions.secret = (.+)",
		).
			WithArgs("123").
			WillReturnRows(row)

		user, err := users.SelectBySession("123")

		if err != nil {
			t.Error(err)
		}

		if expected == nil && user != nil {
			t.Error("Selected user is not nil!")
		} else if !reflect.DeepEqual(user, expected) {
			t.Error("Select by user returned the wrong user!")
		}
	}

	mock.ExpectQuery(
		"SELECT (.+) FROM (.+) WHERE users.id = sessions.user" +
			" AND sessions.secret = (.+)",
	).
		WithArgs("123").
		WillReturnError(errors.New("Dummy error."))

	user, err := users.SelectBySession("123")

	if user != nil {
		t.Error("Select by email should return nil user on error!")
	}

	if err == nil {
		t.Error("Select by email should return non-nil error on error!")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestInsertUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(
		"WITH u AS \\( "+
			"INSERT INTO users (.+) VALUES (.+) RETURNING id "+
			"\\) "+
			"INSERT INTO passwords (.+) "+
			"VALUES (.+)",
	).
		WithArgs("user", "email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(
		`SELECT (.+) FROM (.+) WHERE username = (.+)`,
	).
		WithArgs("user").
		WillReturnRows(testRow())

	users := &Users{db}
	id, err := users.Insert("user", "email@example.com", "p4ssw0rd")

	if err != nil {
		t.Error(err)
	}

	if id != 1 {
		t.Error("Insert with password did not return id 1")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(
		`UPDATE users SET (.+) WHERE id = (.+)`,
	).
		WithArgs("fullname", "new@email.com", "bio", "org", "location", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	users := &Users{db}
	err = users.Update(
		testUser(), "fullname", "new@email.com", "bio", "org", "location",
	)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestVerifyPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	row1 := sqlmock.NewRows([]string{"method", "salt", "hash"}).AddRow(
		"bcrypt",
		"12345",
		"$2a$10$WRv/S87GBhkAkbnnyXD9.ug55t57yai7RD4A/0VdGrgWnOUrE4xJW",
	)
	row2 := sqlmock.NewRows([]string{"method", "salt", "hash"}).AddRow(
		"bcrypt",
		"12345",
		"$2a$10$QbaaWYFip7duAReplLGiHOdFIeHmsqP0TwUhwii00gD9dLKwRzpzi",
	)
	row3 := sqlmock.NewRows([]string{"method", "salt", "hash"})

	trials := map[sqlmock.Rows]bool{
		row1: true,
		row2: false,
		row3: false,
	}

	for row, expected := range trials {
		mock.ExpectQuery(
			`SELECT (.+) FROM passwords, users WHERE (.+)`,
		).
			WithArgs("username").
			WillReturnRows(row)

		users := &Users{db}
		result, err := users.VerifyPassword("username", "password")

		if err == sql.ErrNoRows {

		} else if err != nil {
			t.Error(err)
		}

		if result != expected {
			t.Error("Verify password did not work correctly!")
		}
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}

}

func TestChangePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(
		`UPDATE passwords SET (.+) WHERE user = (.+)`,
	).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	users := &Users{db}
	err = users.ChangePassword(&User{Id: 1}, "newpass")

	if err != nil {
		t.Error(err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestChangePermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(
		`UPDATE users SET (.+) WHERE id = (.+)`,
	).
		WithArgs("VERIFIED", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	users := &Users{db}
	err = users.ChangePermissions(testUser(), "VERIFIED")

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCreateVerificationSecret(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(
		"INSERT INTO user_verifications (.+) VALUES (.+)",
	).
		WithArgs(sqlmock.AnyArg(), "email@example.com", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	users := &Users{db}
	secret, err := users.CreateVerificationSecret(1, "email@example.com")

	if len(secret) == 0 {
		t.Error("Secret is empty!")
	}

	if err != nil {
		t.Error(err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
