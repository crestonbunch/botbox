package api

import (
	"errors"
	"reflect"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNewSessionPutEndpoint(t *testing.T) {
	e := NewSessionPutEndpoint(&App{})
	p := &SessionUpdateProcessors{}

	if e.Path != "/session" {
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
		{e.Processors[1], p.CreateSecret},
	}

	if len(e.Processors) != 2 {
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

func TestSessionPutValidateSession(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Secret string
		User   int
		Rows   sqlmock.Rows
		Err    error
	}

	testCases := map[sampleSetup]*HttpError{
		// good email
		sampleSetup{
			"1234", 1,
			sqlmock.NewRows([]string{"user"}).
				AddRow(1), nil,
		}: nil,
		// strange error
		sampleSetup{
			"1234", 0, nil, errors.New("Dummy Error"),
		}: ErrUnknown,
		// invalid secret
		sampleSetup{
			"1234", 0,
			sqlmock.NewRows([]string{"user"}), nil,
		}: ErrInvalidSecret,
		// empty secret
		sampleSetup{
			"", 0, nil, nil,
		}: ErrInvalidSecret,
	}

	for setup, expected := range testCases {
		model := &SessionPutModel{
			Secret: setup.Secret,
		}

		p := &SessionUpdateProcessors{db: db}

		if setup.Rows != nil {
			mock.ExpectQuery(
				`SELECT \"user\" FROM session_secrets WHERE secret = (.+) ` +
					`AND expires < NOW\(\) AND revoked == FALSE`,
			).
				WithArgs(model.Secret).
				WillReturnRows(setup.Rows)
		} else if setup.Err != nil {
			mock.ExpectQuery(
				`SELECT \"user\" FROM session_secrets WHERE secret = (.+) ` +
					`AND expires < NOW\(\) AND revoked == FALSE`,
			).
				WithArgs(model.Secret).
				WillReturnError(setup.Err)
		}

		m, err := p.ValidateSecret(model)

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

func TestPutSessionCreateSecret(t *testing.T) {
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
		p := SessionUpdateProcessors{db: db, user: test.user}

		q := mock.ExpectExec(
			`INSERT INTO session_secrets \(secret, \"user\"\) VALUES (.+)`,
		).
			WithArgs(sqlmock.AnyArg(), test.user)

		if test.err != nil {
			q.WillReturnError(test.err)
		} else {
			q.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		model := &SessionPutModel{}
		m, err := p.CreateSecret(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}

func TestPutSessionWriter(t *testing.T) {
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

		p := &SessionUpdateProcessors{
			secret: test.Secret,
		}

		w, err := p.Write(&SessionPutModel{})

		if err == test.ExpectErr {
			t.Error(err)
			t.Error("Error did not match expectation!")
		}

		if !reflect.DeepEqual(string(w), string(test.Expect)) {
			t.Errorf("Correct secret %s was not written (%s)!", test.Expect, w)
		}
	}
}
