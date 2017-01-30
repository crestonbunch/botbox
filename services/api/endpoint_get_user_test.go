package api

import (
	"errors"
	"reflect"
	"runtime"
	"testing"

	"database/sql"
	"strconv"
	"time"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNewUserIdGetEndpoint(t *testing.T) {
	e := NewUserIdGetEndpoint(&App{})
	p := &UserIdGetProcessors{}

	if e.Path != "/user/id/{id}" {
		t.Error("Endpoint path is not correct!")
	}

	if !reflect.DeepEqual(e.Methods, []string{"GET"}) {
		t.Error("Endpoint methods are not correct!")
	}

	tests := []struct {
		Source Processor
		Target Processor
	}{
		{e.Processors[0], p.GetUser},
		{e.Processors[1], p.GetPermissions},
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

func TestUserGetIdGetUser(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Id           string
		ReturnRows   sqlmock.Rows
		ReturnErr    error
		ExpectResult *User
		ExpectErr    *HttpError
	}

	cols := []string{"id", "name", "joined", "permission_set"}
	tm := time.Now()
	testCases := []sampleSetup{
		// user not found
		sampleSetup{
			"123",
			nil,
			sql.ErrNoRows,
			nil,
			ErrUserNotFound,
		},
		// strange error
		sampleSetup{
			"123",
			nil,
			errors.New("dummy error"),
			nil,
			ErrUnknown,
		},
		// empty id
		sampleSetup{
			"", nil, nil, nil, ErrMissingParameter,
		},
		// good case
		sampleSetup{
			"123",
			sqlmock.NewRows(cols).AddRow(123, "Joe", tm, "VERIFIED"),
			nil,
			&User{Id: 123, Name: "Joe", Joined: tm, PermissionSet: "VERIFIED"},
			nil,
		},
	}

	for _, test := range testCases {
		model := &UserIdSelectModel{
			Id: test.Id,
		}

		p := &UserIdGetProcessors{db: db}
		id, _ := strconv.Atoi(test.Id)

		if test.ReturnRows != nil {
			mock.ExpectQuery(
				`SELECT id, name, joined, permission_set ` +
					`FROM users WHERE id = (.+)`,
			).
				WithArgs(id).
				WillReturnRows(test.ReturnRows)
		} else if test.ReturnErr != nil {
			mock.ExpectQuery(
				`SELECT id, name, joined, permission_set ` +
					`FROM users WHERE id = (.+)`,
			).
				WithArgs(id).
				WillReturnError(test.ReturnErr)
		}

		m, err := p.GetUser(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err == nil && !reflect.DeepEqual(*p.user, *test.ExpectResult) {
			t.Error("User was not set correctly.")
		}

		if err != test.ExpectErr {
			t.Error(test.ExpectErr)
			t.Error(err)
			t.Error("User get error was unexpected.")
		}

	}
}

func TestUserGetIdGetPermissions(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Set          string
		ReturnRows   sqlmock.Rows
		ReturnErr    error
		ExpectResult PermissionSet
		ExpectErr    *HttpError
	}

	cols := []string{"permission"}
	testCases := []sampleSetup{
		// strange error
		sampleSetup{
			"VERIFIED",
			nil,
			errors.New("dummy error"),
			nil,
			ErrUnknown,
		},
		// good case
		sampleSetup{
			"VERIFIED",
			sqlmock.NewRows(cols).
				AddRow("READ").AddRow("WRITE").AddRow("EXECUTE"),
			nil,
			PermissionSet([]string{"READ", "WRITE", "EXECUTE"}),
			nil,
		},
	}

	for _, test := range testCases {
		model := &UserIdSelectModel{}

		p := &UserIdGetProcessors{db: db, user: &User{PermissionSet: test.Set}}

		if test.ReturnRows != nil {
			mock.ExpectQuery(
				`SELECT permission FROM permission_set_permissions ` +
					`WHERE permission_set = (.+)`,
			).
				WithArgs(test.Set).
				WillReturnRows(test.ReturnRows)
		} else if test.ReturnErr != nil {
			mock.ExpectQuery(
				`SELECT permission FROM permission_set_permissions ` +
					`WHERE permission_set = (.+)`,
			).
				WithArgs(test.Set).
				WillReturnError(test.ReturnErr)
		}

		m, err := p.GetPermissions(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err == nil && !reflect.DeepEqual(p.user.Permissions, test.ExpectResult) {
			t.Error("PermissionSet was not set correctly.")
		}

		if err != test.ExpectErr {
			t.Error(test.ExpectErr)
			t.Error(err)
			t.Error("User get error was unexpected.")
		}

	}
}

func TestUserGetIdWriter(t *testing.T) {

	tests := []struct {
		User *User
		Err  *HttpError
	}{
		{
			User: &User{Id: 123, Name: "Joe", PermissionSet: "VERIFIED"},
			Err:  nil,
		},
	}

	for _, test := range tests {
		model := &UserIdSelectModel{}
		p := &UserIdGetProcessors{user: test.User}

		out, err := p.UserWriter(model)

		if err != test.Err {
			t.Error(err)
		}

		expected, _ := json.Marshal(test.User)

		if string(out) != string(expected) {
			t.Error("Writer returned wrong user")
		}
	}
}
