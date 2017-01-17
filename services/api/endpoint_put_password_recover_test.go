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

func TestNewPasswordRecoverPutEndpoint(t *testing.T) {
	e := NewPasswordRecoverPutEndpoint(&App{})
	p := &PasswordRecoverUpdateProcessors{}

	if e.Path != "/password/recover" {
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
		{e.Processors[1], p.ValidatePassword},
		{e.Processors[2], p.Begin},
		{e.Processors[3], p.UpdateSecret},
		{e.Processors[4], p.UpdatePassword},
		{e.Processors[5], p.Commit},
	}

	if len(e.Processors) != 6 {
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

func TestPasswordRecoverUpdateHandler(t *testing.T) {

	type sampleResponse struct {
		Model *PasswordRecoverPutModel
		Error error
	}

	testCases := map[*http.Request]sampleResponse{
		// Good case
		httptest.NewRequest(
			"PUT", "http://local/password/recover", strings.NewReader(`{
			"secret": "abcdef123456",
			"password": "p455w0rd"
		}`)): sampleResponse{
			Model: &PasswordRecoverPutModel{
				Secret:   "abcdef123456",
				Password: "p455w0rd",
			},
			Error: nil,
		},

		// Invalid JSON
		httptest.NewRequest(
			"PUT", "http://local/password/recover", strings.NewReader(`{
			"secret": "abcdef123456",
			"password": "p455w0rd",
		}`)): sampleResponse{Model: nil, Error: ErrInvalidJson},

		// Invalid reader
		httptest.NewRequest(
			"PUT", "http://local/password/recover",
			iotest.TimeoutReader(strings.NewReader(`{
				"secret": "abcdef123456",
				"password": "p455w0rd"
			}`)),
		): sampleResponse{Model: nil, Error: ErrUnknown},
	}

	for req, expected := range testCases {
		p := &PasswordRecoverUpdateProcessors{}
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
			!reflect.DeepEqual(*expected.Model, *model.(*PasswordRecoverPutModel)) {
			t.Error("Put model loaded the wrong data!")
		}
	}
}

func TestPasswordRecoverValidateSecret(t *testing.T) {
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
		model := &PasswordRecoverPutModel{
			Secret: setup.Secret,
		}

		p := &PasswordRecoverUpdateProcessors{db: db}

		if setup.Rows != nil {
			mock.ExpectQuery(
				"SELECT COUNT\\(secret\\) as count FROM recovery_secrets WHERE secret =" +
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

func TestPasswordRecoverValidatePassword(t *testing.T) {

	testCases := map[string]*HttpError{
		// good password
		"pass1234": nil,
		// empty password
		"": ErrMissingPassword,
		// short password
		"pass": ErrPasswordTooShort,
	}

	for pass, expected := range testCases {
		model := &PasswordRecoverPutModel{
			Password: pass,
		}

		p := &PasswordRecoverUpdateProcessors{}

		m, err := p.ValidatePassword(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if !reflect.DeepEqual(err, expected) {
			t.Error("Password validation error was unexpected.")
		}
	}
}

func TestPasswordRecoverUpdateBegin(t *testing.T) {
	testCases := map[error]*HttpError{
		nil: nil,
		errors.New("Dummy error"): ErrUnknown,
	}
	model := &PasswordRecoverPutModel{
		Secret:   "abcdef123456",
		Password: "p455w0rd",
	}

	for test, expected := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")
		p := PasswordRecoverUpdateProcessors{db: db}

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

func TestPasswordRecoverUpdateSecret(t *testing.T) {
	testCases := []struct {
		model  *PasswordRecoverPutModel
		err    error
		expect error
	}{
		{
			model: &PasswordRecoverPutModel{
				Secret:   "abcdef123456",
				Password: "p455w0rd",
			},
			err:    nil,
			expect: nil,
		},
		{
			model: &PasswordRecoverPutModel{
				Secret:   "abcdef123456",
				Password: "p455w0rd",
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
		p := PasswordRecoverUpdateProcessors{db: db, tx: tx}

		q := mock.ExpectQuery(`UPDATE recovery_secrets SET` +
			` used = TRUE WHERE secret = (.+)` +
			`RETURNING user`,
		).
			WithArgs(test.model.Secret)

		if test.err != nil {
			q.WillReturnError(test.err)
			mock.ExpectRollback()
		} else {
			q.WillReturnRows(sqlmock.NewRows([]string{"user"}).
				AddRow(100))
		}

		m, err := p.UpdateSecret(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}

		if test.err == nil && p.user != 100 {
			t.Error("UpdateSecret did not set the user")
		}
	}
}

func TestPasswordRecoverUpdatePassword(t *testing.T) {
	testCases := []struct {
		model  *PasswordRecoverPutModel
		user   int
		err    error
		expect error
	}{
		{
			model: &PasswordRecoverPutModel{
				Secret:   "abcdef123456",
				Password: "p455w0rd",
			},
			user:   100,
			err:    nil,
			expect: nil,
		},
		{
			model: &PasswordRecoverPutModel{
				Secret:   "abcdef123456",
				Password: "p455w0rd",
			},
			user:   100,
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
		p := PasswordRecoverUpdateProcessors{db: db, tx: tx, user: test.user}

		q := mock.ExpectExec(
			`UPDATE passwords SET hash = (.+), salt = (.+), method = (.+), `+
				`updated = NOW\(\) WHERE user = (.+)`,
		).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "bcrypt", test.user)

		if test.err != nil {
			q.WillReturnError(test.err)
			mock.ExpectRollback()
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

func TestPasswordRecoverUpdateCommit(t *testing.T) {
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
		p := PasswordRecoverUpdateProcessors{db: db, tx: tx}

		begin := mock.ExpectCommit()
		if test != nil {
			begin.WillReturnError(test)
		}
		model := &PasswordRecoverPutModel{
			Secret:   "abcdef123456",
			Password: "p455w0rd",
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
