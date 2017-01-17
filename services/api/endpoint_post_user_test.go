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

func TestNewUserPostEndpoint(t *testing.T) {
	e := NewUserPostEndpoint(&App{})
	p := &UserInsertProcessors{}

	if e.Path != "/user" {
		t.Error("Endpoint path is not correct!")
	}

	if !reflect.DeepEqual(e.Methods, []string{"POST"}) {
		t.Error("Endpoint methods are not correct!")
	}

	tests := []struct {
		Source Processor
		Target Processor
	}{
		{e.Processors[0], p.ValidateName},
		{e.Processors[1], p.ValidateEmail},
		{e.Processors[2], p.ValidatePassword},
		{e.Processors[3], p.ValidateCaptcha},
		{e.Processors[4], p.Begin},
		{e.Processors[5], p.InsertUser},
		{e.Processors[6], p.InsertPassword},
		{e.Processors[7], p.Commit},
	}

	if len(e.Processors) != 8 {
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

func TestUserInsertHandler(t *testing.T) {

	type sampleResponse struct {
		Model *UserPasswordPostModel
		Error error
	}

	testCases := map[*http.Request]sampleResponse{
		// Good case
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"name": "John Doe",
			"password": "pass1234",
			"email": "email@example.com",
			"captcha": "valid"
		}`)): sampleResponse{
			Model: &UserPasswordPostModel{
				Name:     "John Doe",
				Password: "pass1234",
				Email:    "email@example.com",
				Captcha:  "valid",
			},
			Error: nil,
		},

		// Invalid JSON
		httptest.NewRequest(
			"POST", "http://local/user/password", strings.NewReader(`{
			"name": "John Doe",
			"password": "pass1234",
			"email": "email@example.com",
			"captcha": "valid",
		}`)): sampleResponse{Model: nil, Error: ErrInvalidJson},

		// Invalid reader
		httptest.NewRequest(
			"POST", "http://local/user/password",
			iotest.TimeoutReader(strings.NewReader(`{
					"name": "John Doe",
					"password": "pass1234",
					"email": "email@example.com",
					"captcha": "valid"
				}`)),
		): sampleResponse{Model: nil, Error: ErrUnknown},
	}

	for req, expected := range testCases {
		p := &UserInsertProcessors{}
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
			!reflect.DeepEqual(*expected.Model, *model.(*UserPasswordPostModel)) {
			t.Error("Post model loaded the wrong data!")
		}
	}
}

func TestUserInsertValidateEmail(t *testing.T) {
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
		// good email
		sampleSetup{
			"email@example.com", sqlmock.NewRows([]string{"count"}).AddRow(0),
		}: nil,
		// strange error
		sampleSetup{
			"email@example.com", sqlmock.NewRows([]string{"count"}),
		}: ErrUnknown,
		// taken email
		sampleSetup{
			"email@example.com", sqlmock.NewRows([]string{"count"}).AddRow(1),
		}: ErrEmailInUse,
		// empty email
		sampleSetup{
			"", nil,
		}: ErrMissingEmail,
	}

	for setup, expected := range testCases {
		model := &UserPasswordPostModel{
			Name:     "John Doe",
			Password: "pass1234",
			Email:    setup.Email,
			Captcha:  "valid",
		}

		p := &UserInsertProcessors{db: db}

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

func TestUserInsertValidateName(t *testing.T) {

	testCases := map[string]*HttpError{
		// good name
		"John Doe": nil,
		// empty name
		"": ErrMissingName,
		// long name
		"aaaaaaaaaaaaaaaaaaaaa": ErrNameTooLong,
	}

	for name, expected := range testCases {
		model := &UserPasswordPostModel{
			Name:     name,
			Password: "pass1234",
			Email:    "email@example.com",
			Captcha:  "valid",
		}

		p := &UserInsertProcessors{}

		m, err := p.ValidateName(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err != expected {
			t.Error("Name validation error was unexpected.")
		}
	}
}

func TestUserInsertValidatePassword(t *testing.T) {

	testCases := map[string]*HttpError{
		// good password
		"pass1234": nil,
		// empty password
		"": ErrMissingPassword,
		// short password
		"pass": ErrPasswordTooShort,
	}

	for pass, expected := range testCases {
		model := &UserPasswordPostModel{
			Name:     "John Doe",
			Password: pass,
			Email:    "email@example.com",
			Captcha:  "valid",
		}

		p := &UserInsertProcessors{}

		_, err := p.ValidatePassword(model)

		if !reflect.DeepEqual(err, expected) {
			t.Error("Password validation error was unexpected.")
		}
	}
}

type mockRecaptcha struct {
	response bool
	err      error
}

func (r mockRecaptcha) Verify(token string) (bool, error) {
	return r.response, r.err
}

func TestUserInsertValidateCaptcha(t *testing.T) {

	testCases := map[mockRecaptcha]*HttpError{
		// good recaptcha
		mockRecaptcha{true, nil}: nil,
		// bad recaptcha
		mockRecaptcha{false, nil}: ErrBotDetected,
		// unknown error
		mockRecaptcha{false, errors.New("dummy error")}: ErrUnknown,
		// unknown error
		mockRecaptcha{true, errors.New("dummy error")}: ErrUnknown,
	}

	for recaptcha, expected := range testCases {
		model := &UserPasswordPostModel{
			Name:     "John Doe",
			Password: "pass1234",
			Email:    "email@example.com",
			Captcha:  "valid",
		}

		p := &UserInsertProcessors{recaptcha: recaptcha}

		m, err := p.ValidateCaptcha(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err != expected {
			t.Error("Recaptcha validation error was unexpected.")
		}
	}
}

func TestUserInsertBegin(t *testing.T) {
	testCases := map[error]*HttpError{
		nil: nil,
		errors.New("Dummy error"): ErrUnknown,
	}
	model := &UserPasswordPostModel{
		Name:     "John Doe",
		Password: "pass1234",
		Email:    "email@example.com",
		Captcha:  "valid",
	}

	for test, expected := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")
		p := UserInsertProcessors{db: db}

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

func TestUserInsertUser(t *testing.T) {
	testCases := []struct {
		model  *UserPasswordPostModel
		err    error
		expect error
	}{
		{
			model: &UserPasswordPostModel{
				Name:     "John Doe",
				Password: "pass1234",
				Email:    "email@example.com",
				Captcha:  "valid",
			},
			err:    nil,
			expect: nil,
		},
		{
			model: &UserPasswordPostModel{
				Name:     "John Doe",
				Password: "pass1234",
				Email:    "email@example.com",
				Captcha:  "valid",
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
		p := UserInsertProcessors{db: db, tx: tx}

		q := mock.ExpectQuery(`INSERT INTO users `+
			`\(name, email, permission_set\) VALUES (.+) `+
			`RETURNING id`).
			WithArgs(test.model.Name, test.model.Email, "UNVERIFIED")

		if test.err != nil {
			q.WillReturnError(test.err)
			mock.ExpectRollback()
		} else {
			q.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}

		m, err := p.InsertUser(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}

		if test.err == nil && p.userId != 1 {
			t.Error("InsertUser did not set the user id")
		}
	}
}

func TestUserInsertPassword(t *testing.T) {
	testCases := []struct {
		model  *UserPasswordPostModel
		err    error
		expect error
	}{
		{
			model: &UserPasswordPostModel{
				Name:     "John Doe",
				Password: "pass1234",
				Email:    "email@example.com",
				Captcha:  "valid",
			},
			err:    nil,
			expect: nil,
		},
		{
			model: &UserPasswordPostModel{
				Name:     "John Doe",
				Password: "pass1234",
				Email:    "email@example.com",
				Captcha:  "valid",
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
		p := UserInsertProcessors{db: db, tx: tx, userId: 1}

		q := mock.ExpectExec(
			`INSERT INTO passwords \(user, hash, salt, method\) VALUES (.+)`,
		).
			WithArgs(1, sqlmock.AnyArg(), sqlmock.AnyArg(), "bcrypt")

		if test.err != nil {
			q.WillReturnError(test.err)
			mock.ExpectRollback()
		} else {
			q.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		m, err := p.InsertPassword(test.model)

		if err == nil && m != test.model {
			t.Error("Correct model was not returned.")
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	}
}

func TestUserInsertCommit(t *testing.T) {
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
		p := UserInsertProcessors{db: db, tx: tx}

		begin := mock.ExpectCommit()
		if test != nil {
			begin.WillReturnError(test)
		}
		model := &UserPasswordPostModel{
			Name:     "John Doe",
			Password: "pass1234",
			Email:    "email@example.com",
			Captcha:  "valid",
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
