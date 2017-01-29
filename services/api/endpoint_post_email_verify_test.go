package api

import (
	"errors"
	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reflect"
	"runtime"
	"testing"
)

func TestNewEmailVerifyPostEndpoint(t *testing.T) {
	e := NewEmailVerifyPostEndpoint(&App{})
	p := &EmailVerifyInsertProcessors{}

	if e.Path != "/email/verify" {
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
		{e.Processors[2], p.InsertVerification},
		{e.Processors[3], p.SendVerification},
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

func TestEmailVerifyInsertValidate(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Email string
		Name  string
		Rows  sqlmock.Rows
	}

	testCases := map[sampleSetup]*HttpError{
		// good email
		sampleSetup{
			"email@example.com", "John Doe",
			sqlmock.NewRows([]string{"count", "name"}).
				AddRow(1, "John Doe"),
		}: nil,
		// strange error
		sampleSetup{
			"email@example.com", "", sqlmock.NewRows([]string{"count", "name"}),
		}: ErrUnknown,
		// non-existant email
		sampleSetup{
			"email@example.com", "", sqlmock.NewRows([]string{"count", "name"}).
				AddRow(0, ""),
		}: ErrEmailNotFound,
		// empty email
		sampleSetup{
			"", "", nil,
		}: ErrMissingEmail,
	}

	for setup, expected := range testCases {
		model := &EmailVerifyPostModel{
			Email: setup.Email,
		}

		p := &EmailVerifyInsertProcessors{db: db}

		if setup.Rows != nil {
			mock.ExpectQuery(
				"SELECT COUNT\\(id\\) as count, name FROM users WHERE email = (.+)",
			).
				WithArgs(model.Email).
				WillReturnRows(setup.Rows)
		}

		m, err := p.ValidateEmail(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
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

func TestEmailVerifyInsertBegin(t *testing.T) {
	testCases := map[error]*HttpError{
		nil: nil,
		errors.New("Dummy error"): ErrUnknown,
	}
	model := &EmailVerifyPostModel{
		Email: "email@example.com",
	}

	for test, expected := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")
		p := EmailVerifyInsertProcessors{db: db}

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

func TestEmailVerifyInsert(t *testing.T) {
	testCases := []struct {
		model  *EmailVerifyPostModel
		err    error
		expect error
	}{
		{
			model: &EmailVerifyPostModel{
				Email: "email@example.com",
			},
			err:    nil,
			expect: nil,
		},
		{
			model: &EmailVerifyPostModel{
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
		p := EmailVerifyInsertProcessors{db: db, tx: tx}

		q := mock.ExpectExec(
			`INSERT INTO verify_secrets \(secret, email\) VALUES (.+)`,
		).
			WithArgs(sqlmock.AnyArg(), test.model.Email)

		if test.err != nil {
			q.WillReturnError(test.err)
			mock.ExpectRollback()
		} else {
			q.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		m, err := p.InsertVerification(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}

type mockMailer struct {
	err error
}

func (m *mockMailer) SendEmailVerification(email string, name, secret []byte) error {
	return m.err
}
func (m *mockMailer) SendPasswordRecovery(email string, name, secret []byte) error {
	return m.err
}

func TestEmailVerifySend(t *testing.T) {
	testCases := []struct {
		name   string
		model  *EmailVerifyPostModel
		err    error
		expect error
	}{
		{
			name: "John Doe",
			model: &EmailVerifyPostModel{
				Email: "email@example.com",
			},
			err:    nil,
			expect: nil,
		},
		{
			name: "John Doe",
			model: &EmailVerifyPostModel{
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

		p := EmailVerifyInsertProcessors{
			db: db, tx: tx, name: test.name, emailer: &mockMailer{test.err},
		}

		if test.expect != nil {
			mock.ExpectRollback()
		}

		m, err := p.SendVerification(test.model)

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

func TestEmailVerifyInsertCommit(t *testing.T) {
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
		p := EmailVerifyInsertProcessors{db: db, tx: tx}

		begin := mock.ExpectCommit()
		if test != nil {
			begin.WillReturnError(test)
		}
		model := &EmailVerifyPostModel{
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
