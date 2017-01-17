package api

import (
	"errors"
	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"
)

func TestNewPasswordRecoverPostEndpoint(t *testing.T) {
	e := NewPasswordRecoverPostEndpoint(&App{})
	p := &PasswordRecoverInsertProcessers{}

	if e.Path != "/password/recover" {
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
		{e.Processors[1], p.Begin},
		{e.Processors[2], p.InsertRecovery},
		{e.Processors[3], p.SendRecovery},
		{e.Processors[4], p.Commit},
	}

	if len(e.Processors) != 5 {
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

func TestPasswordRecoverInsertHandler(t *testing.T) {

	type sampleResponse struct {
		Model *PasswordRecoverPostModel
		Error error
	}

	testCases := map[*http.Request]sampleResponse{
		// Good case
		httptest.NewRequest(
			"POST", "http://local/password/recover", strings.NewReader(`{
			"email": "email@example.com"
		}`)): sampleResponse{
			Model: &PasswordRecoverPostModel{
				Email: "email@example.com",
			},
			Error: nil,
		},

		// Invalid JSON
		httptest.NewRequest(
			"POST", "http://local/password/recover", strings.NewReader(`{
			"email": "email@example.com",
		}`)): sampleResponse{Model: nil, Error: ErrInvalidJson},

		// Invalid reader
		httptest.NewRequest(
			"POST", "http://local/password/recover",
			iotest.TimeoutReader(strings.NewReader(`{
					"email": "email@example.com"
				}`)),
		): sampleResponse{Model: nil, Error: ErrUnknown},
	}

	for req, expected := range testCases {
		p := &PasswordRecoverInsertProcessers{}
		model, err := p.Handler(req)

		if err != nil && expected.Error == nil {
			t.Error(err)
		} else if err == nil && expected.Error != nil {
			t.Error("Post model returned a nil error!")
		} else if expected.Error == nil && err == nil {
			// Why does deep equal not correctly compare nil and nil?
		} else if !reflect.DeepEqual(expected.Error, err) {
			t.Error("Post model returned the wrong error!")
		}

		if expected.Model != nil &&
			!reflect.DeepEqual(*expected.Model, *model.(*PasswordRecoverPostModel)) {
			t.Error("Post model loaded the wrong data!")
		}
	}
}

func TestPasswordRecoverValidateEmail(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Email string
		Name  string
		User  int
		Rows  sqlmock.Rows
	}

	testCases := map[sampleSetup]*HttpError{
		// good email
		sampleSetup{
			"email@example.com", "John Doe", 1,
			sqlmock.NewRows([]string{"count", "name", "id"}).
				AddRow(1, "John Doe", 1),
		}: nil,
		// strange error
		sampleSetup{
			"email@example.com", "", 0,
			sqlmock.NewRows([]string{"count", "name", "id"}),
		}: ErrUnknown,
		// non-existant email
		sampleSetup{
			"email@example.com", "", 0,
			sqlmock.NewRows([]string{"count", "name", "id"}).
				AddRow(0, "", 0),
		}: ErrEmailNotFound,
		// empty email
		sampleSetup{
			"", "", 0, nil,
		}: ErrMissingEmail,
	}

	for setup, expected := range testCases {
		model := &PasswordRecoverPostModel{
			Email: setup.Email,
		}

		p := &PasswordRecoverInsertProcessers{db: db}

		if setup.Rows != nil {
			mock.ExpectQuery(
				"SELECT COUNT\\(id\\) as count, name, id FROM users WHERE email = (.+)",
			).
				WithArgs(model.Email).
				WillReturnRows(setup.Rows)
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
		} else if err == nil && p.name != setup.Name {
			t.Log(p.name)
			t.Error("Expected name was not correct")
		}

	}
}

func TestPasswordRecoverBegin(t *testing.T) {
	testCases := map[error]*HttpError{
		nil: nil,
		errors.New("Dummy error"): ErrUnknown,
	}
	model := &PasswordRecoverPostModel{
		Email: "email@example.com",
	}

	for test, expected := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")
		p := PasswordRecoverInsertProcessers{db: db}

		begin := mock.ExpectBegin()
		if test != nil {
			begin.WillReturnError(test)
		}

		m, err := p.Begin(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if !reflect.DeepEqual(err, expected) {
			t.Error("Begin returned the wrong expected error.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}

		if expected == nil && p.Begin == nil {
			t.Error("Begin did not set the transaction.")
		}
	}
}

func TestPasswordRecoverInsert(t *testing.T) {
	testCases := []struct {
		model  *PasswordRecoverPostModel
		user   int
		err    error
		expect error
	}{
		{
			model: &PasswordRecoverPostModel{
				Email: "email@example.com",
			},
			user:   1,
			err:    nil,
			expect: nil,
		},
		{
			model: &PasswordRecoverPostModel{
				Email: "email@example.com",
			},
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
		mock.ExpectBegin()
		tx, err := db.Beginx()
		if err != nil {
			t.Error(err)
		}
		p := PasswordRecoverInsertProcessers{db: db, tx: tx, user: test.user}

		q := mock.ExpectExec(
			`INSERT INTO recovery_secrets \(secret, user\) VALUES (.+)`,
		).
			WithArgs(sqlmock.AnyArg(), test.user)

		if test.err != nil {
			q.WillReturnError(test.err)
			mock.ExpectRollback()
		} else {
			q.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		m, err := p.InsertRecovery(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}

func TestPasswordRecoverSend(t *testing.T) {
	testCases := []struct {
		name   string
		model  *PasswordRecoverPostModel
		err    error
		expect error
	}{
		{
			name: "John Doe",
			model: &PasswordRecoverPostModel{
				Email: "email@example.com",
			},
			err:    nil,
			expect: nil,
		},
		{
			name: "John Doe",
			model: &PasswordRecoverPostModel{
				Email: "email@example.com",
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
		mock.ExpectBegin()
		tx, err := db.Beginx()
		if err != nil {
			t.Error(err)
		}

		p := PasswordRecoverInsertProcessers{
			db: db, tx: tx, name: test.name, emailer: &mockMailer{test.err},
		}

		if test.expect != nil {
			mock.ExpectRollback()
		}

		m, err := p.SendRecovery(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if test.expect != nil && err != nil && test.expect != err {
			t.Error(err)
			t.Error("Error did not match expectation.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}

func TestPasswordRecoverCommit(t *testing.T) {
	testCases := map[error]*HttpError{
		nil: nil,
		errors.New("Dummy error"): ErrUnknown,
	}

	for test, expected := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")
		mock.ExpectBegin()
		tx, err := db.Beginx()
		if err != nil {
			t.Error(err)
		}
		p := PasswordRecoverInsertProcessers{db: db, tx: tx}

		begin := mock.ExpectCommit()
		if test != nil {
			begin.WillReturnError(test)
		}
		model := &PasswordRecoverPostModel{
			Email: "email@example.com",
		}

		m, err := p.Commit(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if !reflect.DeepEqual(err, expected) {
			t.Error("Commit returned the wrong expected error.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}

}
