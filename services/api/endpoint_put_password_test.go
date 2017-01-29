package api

import (
	"errors"
	"reflect"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNewPasswordPutEndpoint(t *testing.T) {
	e := NewPasswordPutEndpoint(&App{})
	p := &PasswordUpdateProcessors{}

	if e.Path != "/password" {
		t.Error("Endpoint path is not correct!")
	}

	if !reflect.DeepEqual(e.Methods, []string{"PUT"}) {
		t.Error("Endpoint methods are not correct!")
	}

	tests := []struct {
		Source Processor
		Target Processor
	}{
		{e.Processors[0], p.CheckPasswordMatch},
		{e.Processors[1], p.ValidatePassword},
		{e.Processors[2], p.UpdatePassword},
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

func TestPasswordPutPasswordMatch(t *testing.T) {
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
			ExpectErr: ErrInvalidPassword,
		},
		// error test
		sampleSetup{
			TestPassword: "password",
			TestRows:     sqlmock.NewRows([]string{"user", "method", "hash", "salt"}),
			TestErr:      errors.New("dummy"),
			ExpectErr:    ErrInvalidPassword,
		},
	}

	for _, setup := range testCases {
		model := &PasswordPutModel{
			Old: setup.TestPassword,
			New: "blah",
		}

		p := &PasswordUpdateProcessors{
			db:      db,
			handler: &JsonHandlerWithAuth{User: 101},
		}

		if setup.TestRows != nil {
			mock.ExpectQuery(
				`SELECT user, method, hash, salt FROM passwords WHERE \"user\" = (.+)`,
			).
				WithArgs(101).
				WillReturnRows(setup.TestRows)
		} else if setup.TestErr != nil {
			mock.ExpectQuery(
				`SELECT user, method, hash, salt FROM passwords WHERE \"user\" = (.+)`,
			).WithArgs(101).WillReturnError(setup.TestErr)
		}

		m, err := p.CheckPasswordMatch(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err != setup.ExpectErr {
			t.Error(err)
			t.Error("Password matching error was unexpected.")
		}
	}
}

func TestUpdatePasswordValidatePassword(t *testing.T) {

	testCases := map[string]*HttpError{
		// good password
		"pass1234": nil,
		// empty password
		"": ErrMissingPassword,
		// short password
		"pass": ErrPasswordTooShort,
	}

	for pass, expected := range testCases {
		model := &PasswordPutModel{
			New: pass,
		}

		p := &PasswordUpdateProcessors{}

		_, err := p.ValidatePassword(model)

		if !reflect.DeepEqual(err, expected) {
			t.Error("Password validation error was unexpected.")
		}
	}
}

func TestUpdatePasswordUpdatePassword(t *testing.T) {
	testCases := []struct {
		model  *PasswordPutModel
		err    error
		expect error
	}{
		{
			model: &PasswordPutModel{
				Old: "abcd12345",
				New: "p455w0rd",
			},
			err:    nil,
			expect: nil,
		},
		{
			model: &PasswordPutModel{
				Old: "abcd12345",
				New: "p455w0rd",
			},
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
		p := &PasswordUpdateProcessors{
			db:      db,
			handler: &JsonHandlerWithAuth{User: 101},
		}

		q := mock.ExpectExec(`UPDATE passwords \(salt, hash, method\) `+
			`VALUES (.+) WHERE \"user\" = (.+)`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "bcrypt", 101)

		if test.err != nil {
			q.WillReturnError(test.err)
		} else {
			q.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		m, err := p.UpdatePassword(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}
