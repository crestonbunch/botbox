package api

import (
	"errors"
	"reflect"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNewSessionPostEndpoint(t *testing.T) {
	e := NewSessionPostEndpoint(&App{})
	p := &SessionInsertProcessors{}

	if e.Path != "/session" {
		t.Error("Endpoint path is not correct!")
	}

	if !reflect.DeepEqual(e.Methods, []string{"POST"}) {
		t.Error("Endpoint methods are not correct!")
	}

	tests := []struct {
		Source Processor
		Target Processor
	}{
		{e.Processors[0], p.ValidateEmail},
		{e.Processors[1], p.ValidatePassword},
		{e.Processors[2], p.CreateSecret},
	}

	if len(e.Processors) != 3 {
		t.Error("Incorrect number of processors.")
	}

	for _, test := range tests {
		val := runtime.FuncForPC(reflect.ValueOf(test.Source).Pointer()).Name()
		exp := runtime.FuncForPC(reflect.ValueOf(test.Target).Pointer()).Name()

		if val != exp {
			t.Error("Processors in unexpected order.")
		}
	}
}

func TestSessionPostValidateEmail(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Email string
		User  int
		Rows  sqlmock.Rows
		Err   error
	}

	testCases := map[sampleSetup]*HttpError{
		// good email
		sampleSetup{
			"email@example.com", 1,
			sqlmock.NewRows([]string{"id"}).
				AddRow(1), nil,
		}: nil,
		// strange error
		sampleSetup{
			"email@example.com", 0, nil, errors.New("Dummy Error"),
		}: ErrUnknown,
		// non-existant email
		sampleSetup{
			"email@example.com", 0,
			sqlmock.NewRows([]string{"id"}), nil,
		}: ErrLoginIncorrect,
		// empty email
		sampleSetup{
			"", 0, nil, nil,
		}: ErrMissingEmail,
	}

	for setup, expected := range testCases {
		model := &SessionPostModel{
			Email: setup.Email,
		}

		p := &SessionInsertProcessors{db: db}

		if setup.Rows != nil {
			mock.ExpectQuery(
				"SELECT id FROM users WHERE email = (.+)",
			).
				WithArgs(model.Email).
				WillReturnRows(setup.Rows)
		} else if setup.Err != nil {
			mock.ExpectQuery(
				"SELECT id FROM users WHERE email = (.+)",
			).
				WithArgs(model.Email).
				WillReturnError(setup.Err)
		}

		m, err := p.ValidateEmail(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if p.user != setup.User {
			t.Error("User was not set correctly.")
		}

		if err != expected {
			t.Error(err)
			t.Error("Email validation error was unexpected.")
		}
	}
}

func TestPostSessionValidatePassword(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		TestPassword string
		TestRows     sqlmock.Rows
		TestErr      error
		ExpectErr    *HttpError
	}

	testCases := []sampleSetup{
		// good test
		sampleSetup{
			TestPassword: "p455w0rd",
			TestRows: sqlmock.NewRows([]string{"user", "method", "hash", "salt"}).
				AddRow(
					101,
					"bcrypt",
					"$2a$10$RLHXJkb2BFeEHwhONKnBPeoj4O2T5SiJmKAy9TCpK4lTcVonqT8TO",
					"1234",
				),
			TestErr:   nil,
			ExpectErr: nil,
		},
		// bad test
		sampleSetup{
			TestPassword: "password",
			TestRows: sqlmock.NewRows([]string{"user", "method", "hash", "salt"}).
				AddRow(
					101,
					"bcrypt",
					"$2a$10$RLHXJkb2BFeEHwhONKnBPeoj4O2T5SiJmKAy9TCpK4lTcVonqT8TO",
					"1234",
				),
			TestErr:   nil,
			ExpectErr: ErrLoginIncorrect,
		},
		// error test
		sampleSetup{
			TestPassword: "password",
			TestRows:     sqlmock.NewRows([]string{"user", "method", "hash", "salt"}),
			TestErr:      errors.New("dummy"),
			ExpectErr:    ErrLoginIncorrect,
		},
	}

	for _, setup := range testCases {
		model := &SessionPostModel{
			Password: setup.TestPassword,
		}

		p := &SessionInsertProcessors{
			db:   db,
			user: 101,
		}

		if setup.TestRows != nil {
			mock.ExpectQuery(
				`SELECT \"user\", method, hash, salt FROM passwords WHERE \"user\" = (.+)`,
			).
				WithArgs(101).
				WillReturnRows(setup.TestRows)
		} else if setup.TestErr != nil {
			mock.ExpectQuery(
				`SELECT \"user\", method, hash, salt FROM passwords WHERE \"user\" = (.+)`,
			).WithArgs(101).WillReturnError(setup.TestErr)
		}

		m, err := p.ValidatePassword(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err != setup.ExpectErr {
			t.Error(err)
			t.Error("Password matching error was unexpected.")
		}
	}
}

func TestPostSessionCreateSecret(t *testing.T) {
	testCases := []struct {
		user   int
		err    error
		expect error
	}{
		{
			user:   1,
			err:    nil,
			expect: nil,
		},
		{
			user:   0,
			err:    errors.New("Dummy error"),
			expect: ErrUnknown,
		},
	}

	for _, test := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")
		if err != nil {
			t.Error(err)
		}
		p := SessionInsertProcessors{db: db, user: test.user}

		q := mock.ExpectExec(
			`INSERT INTO session_secrets \(secret, \"user\"\) VALUES (.+)`,
		).
			WithArgs(sqlmock.AnyArg(), test.user)

		if test.err != nil {
			q.WillReturnError(test.err)
		} else {
			q.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		model := &SessionPostModel{}
		m, err := p.CreateSecret(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}

func TestPostSessionWriter(t *testing.T) {
	testCases := []struct {
		Secret    string
		Expect    []byte
		ExpectErr error
	}{
		{
			Secret:    "1234",
			Expect:    []byte("1234"),
			ExpectErr: nil,
		},
	}

	for _, test := range testCases {

		p := &SessionInsertProcessors{
			secret: test.Secret,
		}

		w, err := p.Write(&SessionPostModel{})

		if err == test.ExpectErr {
			t.Error(err)
			t.Error("Error did not match expectation!")
		}

		if !reflect.DeepEqual(string(w), string(test.Expect)) {
			t.Errorf("Correct secret %s was not written (%s)!", test.Expect, w)
		}
	}
}
