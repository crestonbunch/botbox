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

func TestNewEmailVerifyPutEndpoint(t *testing.T) {
	e := NewEmailVerifyPutEndpoint(&App{})
	p := &EmailVerifyUpdateProcessors{}

	if e.Path != "/email/verify" {
		t.Error("Endpoint path is not correct!")
	}

	if !reflect.DeepEqual(e.Methods, []string{"PUT"}) {
		t.Error("Endpoint methods are not correct!")
	}

	tests := []struct {
		Source Processor
		Target Processor
	}{
		{e.Processors[0], p.ValidateSecret},
		{e.Processors[1], p.Begin},
		{e.Processors[2], p.UpdateSecret},
		{e.Processors[3], p.UpdateUser},
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

func TestVerifyEmailUpdateHandler(t *testing.T) {

	type sampleResponse struct {
		Model *EmailVerifyPutModel
		Error error
	}

	testCases := map[*http.Request]sampleResponse{
		// Good case
		httptest.NewRequest(
			"PUT", "http://local/email/verify", strings.NewReader(`{
			"secret": "abcdef123456"
		}`)): sampleResponse{
			Model: &EmailVerifyPutModel{
				Secret: "abcdef123456",
			},
			Error: nil,
		},

		// Invalid JSON
		httptest.NewRequest(
			"PUT", "http://local/email/verify", strings.NewReader(`{
			"secret": "abcdef123456",
		}`)): sampleResponse{Model: nil, Error: ErrInvalidJson},

		// Invalid reader
		httptest.NewRequest(
			"PUT", "http://local/email/verify",
			iotest.TimeoutReader(strings.NewReader(`{
				"secret": "abcdef123456"
			}`)),
		): sampleResponse{Model: nil, Error: ErrUnknown},
	}

	for req, expected := range testCases {
		p := &EmailVerifyUpdateProcessors{}
		model, err := p.Handler(req)

		if err != nil && expected.Error == nil {
			t.Error(err)
		} else if err == nil && expected.Error != nil {
			t.Error("Put model returned a nil error!")
		} else if expected.Error == nil && err == nil {
			// Why does deep equal not correctly compare nil and nil?
		} else if !reflect.DeepEqual(expected.Error, err) {
			t.Error("Put model returned the wrong error!")
		}

		if expected.Model != nil &&
			!reflect.DeepEqual(*expected.Model, *model.(*EmailVerifyPutModel)) {
			t.Error("Put model loaded the wrong data!")
		}
	}
}

func TestEmailVerifyValidateSecret(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Secret string
		Rows   sqlmock.Rows
	}

	testCases := map[sampleSetup]*HttpError{
		// good secret
		sampleSetup{
			"abcdef12346", sqlmock.NewRows([]string{"count"}).AddRow(1),
		}: nil,
		// strange error
		sampleSetup{
			"abcdef123456", sqlmock.NewRows([]string{"count"}),
		}: ErrUnknown,
		// no such secret
		sampleSetup{
			"abcdef123456", sqlmock.NewRows([]string{"count"}).AddRow(0),
		}: ErrInvalidSecret,
		// empty secret
		sampleSetup{
			"", nil,
		}: ErrInvalidSecret,
	}

	for setup, expected := range testCases {
		model := &EmailVerifyPutModel{
			Secret: setup.Secret,
		}

		p := &EmailVerifyUpdateProcessors{db: db}

		if setup.Rows != nil {
			mock.ExpectQuery(
				"SELECT COUNT\\(secret\\) as count FROM verify_secrets WHERE secret =" +
					" (.+) AND expires > NOW\\(\\) AND used == FALSE",
			).
				WithArgs(model.Secret).
				WillReturnRows(setup.Rows)
		}

		m, err := p.ValidateSecret(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err != expected {
			t.Error(err)
			t.Error("Secret validation error was unexpected.")
		}

	}
}

func TestEmailVerifyUpdateBegin(t *testing.T) {
	testCases := map[error]*HttpError{
		nil: nil,
		errors.New("Dummy error"): ErrUnknown,
	}
	model := &EmailVerifyPutModel{
		Secret: "abcdef123456",
	}

	for test, expected := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")
		p := EmailVerifyUpdateProcessors{db: db}

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

func TestEmailVerifyUpdateSecret(t *testing.T) {
	testCases := []struct {
		model  *EmailVerifyPutModel
		err    error
		expect error
	}{
		{
			model: &EmailVerifyPutModel{
				Secret: "abcdef123456",
			},
			err:    nil,
			expect: nil,
		},
		{
			model: &EmailVerifyPutModel{
				Secret: "abcdef123456",
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
		p := EmailVerifyUpdateProcessors{db: db, tx: tx}

		q := mock.ExpectQuery(`UPDATE verify_secrets SET` +
			` used = TRUE WHERE secret = (.+)` +
			`RETURNING email`,
		).
			WithArgs(test.model.Secret)

		if test.err != nil {
			q.WillReturnError(test.err)
			mock.ExpectRollback()
		} else {
			q.WillReturnRows(sqlmock.NewRows([]string{"email"}).
				AddRow("email@example.com"))
		}

		m, err := p.UpdateSecret(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}

		if test.err == nil && p.email != "email@example.com" {
			t.Error("UpdateSecret did not set the email")
		}
	}
}

func TestEmailVerifyUpdateUser(t *testing.T) {
	testCases := []struct {
		model  *EmailVerifyPutModel
		email  string
		err    error
		expect error
	}{
		{
			model: &EmailVerifyPutModel{
				Secret: "abcdef123456",
			},
			email:  "email@example.com",
			err:    nil,
			expect: nil,
		},
		{
			model: &EmailVerifyPutModel{
				Secret: "abcdef123456",
			},
			email:  "email@example.com",
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
		p := EmailVerifyUpdateProcessors{db: db, tx: tx, email: test.email}

		q := mock.ExpectExec(
			`UPDATE users SET permission_set = (.+) WHERE email = (.+) AND`+
				` permission_set = (.+)`,
		).
			WithArgs("VERIFIED", p.email, "UNVERIFIED")

		if test.err != nil {
			q.WillReturnError(test.err)
			mock.ExpectRollback()
		} else {
			q.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		m, err := p.UpdateUser(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}

func TestEmailVerifyUpdateCommit(t *testing.T) {
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
		p := EmailVerifyUpdateProcessors{db: db, tx: tx}

		begin := mock.ExpectCommit()
		if test != nil {
			begin.WillReturnError(test)
		}
		model := &EmailVerifyPutModel{
			Secret: "abcdef123456",
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
