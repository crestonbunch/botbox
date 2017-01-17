package api

import (
	"errors"
	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reflect"
	"testing"
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
				`SELECT id FROM users, session_secrets ` +
					`WHERE secret = (.+) AND expires > NOW\(\) AND used == FALSE ` +
					`AND user = id`,
			).
				WithArgs(test.Secret).
				WillReturnError(test.ResultErr)
		} else {
			mock.ExpectQuery(
				`SELECT id FROM users, session_secrets ` +
					`WHERE secret = (.+) AND expires > NOW\(\) AND used == FALSE ` +
					`AND user = id`,
			).
				WithArgs(test.Secret).
				WillReturnRows(test.ResultRows)
		}

		id, err := testSession.GetUserId(test.Secret)

		if id != test.ExpectId {
			t.Error("Returned wrong id!")
		}

		if err != test.ExpectErr {
			t.Error("Returned wrong error!")
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
				`SELECT permission FROM permission_set_permissions, users, session_secrets ` +
					`WHERE secret = (.+) AND users.id = session_secrets\.user ` +
					`AND permission_set_permissions\.permission_set = users\.permission_set`,
			).
				WithArgs(test.Secret).
				WillReturnError(test.ResultErr)
		} else {
			mock.ExpectQuery(
				`SELECT permission FROM permission_set_permissions, users, session_secrets ` +
					`WHERE secret = (.+) AND users.id = session_secrets\.user ` +
					`AND permission_set_permissions\.permission_set = users\.permission_set`,
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
