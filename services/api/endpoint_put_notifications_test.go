package api

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNewNotificationsPutEndpoint(t *testing.T) {
	e := NewNotificationsPutEndpoint(&App{})
	p := &NotificationsUpdateProcessors{}

	if e.Path != "/notifications" {
		t.Error("Endpoint path is not correct!")
	}

	if !reflect.DeepEqual(e.Methods, []string{"PUT"}) {
		t.Error("Endpoint methods are not correct!")
	}

	tests := []struct {
		Source Processor
		Target Processor
	}{
		{e.Processors[0], p.UpdateNotifications},
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

func TestUpdateNotifications(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type TestSetup struct {
		InputModel   NotificationsPutModel
		ReturnResult sql.Result
		ReturnErr    error
		ExpectErr    error
	}

	tests := []TestSetup{
		// good
		TestSetup{
			InputModel: NotificationsPutModel{
				Notifications: []int{1, 2, 3, 4},
				Read:          true,
				Dismissed:     false,
			},
			ReturnResult: sqlmock.NewResult(1, 4),
			ReturnErr:    nil,
			ExpectErr:    nil,
		},
		// good
		TestSetup{
			InputModel: NotificationsPutModel{
				Notifications: []int{1, 2, 3, 4},
				Read:          false,
				Dismissed:     true,
			},
			ReturnResult: sqlmock.NewResult(1, 4),
			ReturnErr:    nil,
			ExpectErr:    nil,
		},
		// strange error
		TestSetup{
			InputModel: NotificationsPutModel{
				Notifications: []int{1, 2, 3, 4},
				Read:          true,
				Dismissed:     true,
			},
			ReturnResult: nil,
			ReturnErr:    errors.New("dummy"),
			ExpectErr:    ErrUnknown,
		},
		// both false
		TestSetup{
			InputModel: NotificationsPutModel{
				Notifications: []int{1, 2, 3, 4},
				Read:          false,
				Dismissed:     false,
			},
			ReturnResult: nil,
			ReturnErr:    nil,
			ExpectErr:    nil,
		},
		// no notifications
		TestSetup{
			InputModel: NotificationsPutModel{
				Notifications: []int{},
				Read:          true,
				Dismissed:     true,
			},
			ReturnResult: nil,
			ReturnErr:    nil,
			ExpectErr:    ErrMissingNotifications,
		},
	}

	for _, setup := range tests {
		model := &setup.InputModel

		p := &NotificationsUpdateProcessors{
			db:      db,
			handler: &JsonHandlerWithAuth{User: 123},
		}

		set := ""
		if model.Read && !model.Dismissed {
			set = `read = NOW\(\)`
		} else if model.Dismissed && !model.Read {
			set = `dismissed = NOW\(\)`
		} else if model.Dismissed && model.Read {
			set = `read = NOW\(\), dismissed = NOW\(\)`
		}
		args := []driver.Value{}
		args = append(args, p.handler.User)
		for _, n := range setup.InputModel.Notifications {
			args = append(args, n)
		}

		if setup.ReturnResult != nil {
			mock.ExpectExec(`UPDATE notifications ` +
				`SET ` + set +
				` WHERE "user" = (.+) AND id IN \((.+)\)`).
				WithArgs(args...).
				WillReturnResult(setup.ReturnResult)
		} else if setup.ReturnErr != nil {
			mock.ExpectExec(`UPDATE notifications ` +
				`SET read = (.+), dismissed = (.+) ` +
				`WHERE "user" = (.+) AND id IN \((.+)\)`).
				WithArgs(args...).
				WillReturnError(setup.ReturnErr)
		}

		m, err := p.UpdateNotifications(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}

		if err != nil && setup.ExpectErr != nil {
			if err != setup.ExpectErr {
				t.Errorf("%s was unexpected, expected %s", err, setup.ExpectErr)
			}
		}
	}

}
