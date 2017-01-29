package api

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNewEmailGetEndpoint(t *testing.T) {
	e := NewEmailGetEndpoint(&App{})
	p := &EmailSelectProcessors{}

	if e.Path != "/email/{email}" {
		t.Error("Endpoint path is not correct!")
	}

	if !reflect.DeepEqual(e.Methods, []string{"GET"}) {
		t.Error("Endpoint methods are not correct!")
	}

	tests := []struct {
		Source Processor
		Target Processor
	}{
		{e.Processors[0], p.ValidateEmail},
	}

	if len(e.Processors) != 1 {
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

func TestNewEmailGetValidateEmail(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Email string
		Rows  sqlmock.Rows
	}

	testCases := map[sampleSetup]*HttpError{
		// email not found
		sampleSetup{
			"email@example.com", sqlmock.NewRows([]string{"count"}).AddRow(0),
		}: ErrEmailNotFound,
		// strange error
		sampleSetup{
			"email@example.com", sqlmock.NewRows([]string{"count"}),
		}: ErrUnknown,
		// email is in use
		sampleSetup{
			"email@example.com", sqlmock.NewRows([]string{"count"}).AddRow(1),
		}: nil,
		// empty email
		sampleSetup{
			"", nil,
		}: ErrMissingEmail,
	}

	for setup, expected := range testCases {
		model := &EmailGetModel{
			Email: setup.Email,
		}

		p := &EmailSelectProcessors{db: db}

		if setup.Rows != nil {
			mock.ExpectQuery(
				"SELECT COUNT\\(id\\) as count FROM users WHERE email = (.+)",
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
		}

	}
}
