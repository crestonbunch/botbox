package api

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetUserId(t *testing.T) {
	dummyErr := errors.New("Dummy error")
	testCases := []struct {
		Secret     string
		ResultRows sqlmock.Rows
		ResultErr  error
		ExpectId   int
		ExpectErr  error
	}{
		{
			Secret:     "abcd1234",
			ResultRows: sqlmock.NewRows([]string{"id"}).AddRow(101),
			ResultErr:  nil,
			ExpectId:   101,
			ExpectErr:  nil,
		},
		{
			Secret:     "abcd1234",
			ResultRows: sqlmock.NewRows([]string{"id"}),
			ResultErr:  dummyErr,
			ExpectId:   0,
			ExpectErr:  dummyErr,
		},
	}

	for _, test := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")

		testSession := &Session{db: db}

		if test.ResultErr != nil {
			mock.ExpectQuery(
				`SELECT id FROM users INNER JOIN session_secrets ` +
					`ON \(users.id = session_secrets.user\) ` +
					`WHERE secret = (.+) AND expires > NOW\(\) AND revoked = FALSE`,
			).
				WithArgs(test.Secret).
				WillReturnError(test.ResultErr)
		} else {
			mock.ExpectQuery(
				`SELECT id FROM users INNER JOIN session_secrets ` +
					`ON \(users.id = session_secrets.user\) ` +
					`WHERE secret = (.+) AND expires > NOW\(\) AND revoked = FALSE`,
			).
				WithArgs(test.Secret).
				WillReturnRows(test.ResultRows)
		}

		id, err := testSession.GetUserId(test.Secret)

		if id != test.ExpectId {
			t.Errorf("Expected id %d got %d", test.ExpectId, id)
		}

		if err != test.ExpectErr {
			t.Errorf("Expected err '%s' got '%s'\n", test.ExpectErr, err)
		}
	}
}

func TestGetPermissions(t *testing.T) {
	dummyErr := errors.New("Dummy error")
	testCases := []struct {
		Secret            string
		ResultRows        sqlmock.Rows
		ResultErr         error
		ExpectPermissions []string
		ExpectErr         error
	}{
		{
			Secret: "abcd1234",
			ResultRows: sqlmock.NewRows([]string{"permission"}).
				AddRow("POST_COMMENT").AddRow("EDIT_PROFILE").AddRow("UPLOAD_FILE"),
			ResultErr:         nil,
			ExpectPermissions: []string{"POST_COMMENT", "EDIT_PROFILE", "UPLOAD_FILE"},
			ExpectErr:         nil,
		},
		{
			Secret:            "abcd1234",
			ResultRows:        sqlmock.NewRows([]string{"permission"}),
			ResultErr:         dummyErr,
			ExpectPermissions: nil,
			ExpectErr:         dummyErr,
		},
	}

	for _, test := range testCases {
		mockDb, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(mockDb, "sqlmock")

		testSession := &Session{db: db}

		if test.ResultErr != nil {
			mock.ExpectQuery(
				`SELECT permission FROM permission_set_permissions ` +
					`INNER JOIN users ` +
					`ON \(users.permission_set = permission_set_permissions.permission_set\) ` +
					`INNER JOIN session_secrets ` +
					`ON \(session_secrets.user = users.id\) ` +
					`WHERE session_secrets.secret = (.+) ` +
					`AND users.id = session_secrets.user ` +
					`AND permission_set_permissions.permission_set = users.permission_set`,
			).
				WithArgs(test.Secret).
				WillReturnError(test.ResultErr)
		} else {
			mock.ExpectQuery(
				`SELECT permission FROM permission_set_permissions ` +
					`INNER JOIN users ` +
					`ON \(users.permission_set = permission_set_permissions.permission_set\) ` +
					`INNER JOIN session_secrets ` +
					`ON \(session_secrets.user = users.id\) ` +
					`WHERE session_secrets.secret = (.+) ` +
					`AND users.id = session_secrets.user ` +
					`AND permission_set_permissions.permission_set = users.permission_set`,
			).
				WithArgs(test.Secret).
				WillReturnRows(test.ResultRows)
		}

		permissions, err := testSession.GetPermissions(test.Secret)

		if reflect.DeepEqual(permissions, test.ExpectPermissions) {
			t.Error("Returned wrong id!")
		}

		if err != test.ExpectErr {
			t.Error(err)
			t.Error("Returned wrong error!")
		}
	}
}
