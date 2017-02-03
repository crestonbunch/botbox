package api

import (
	"encoding/json"
	"errors"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNewNotificationsGetEndpoint(t *testing.T) {
	e := NewNotificationsGetEndpoint(&App{})
	p := &NotificationsSelectProcessors{}

	if e.Path != "/notifications" {
		t.Error("Endpoint path is not correct!")
	}

	if !reflect.DeepEqual(e.Methods, []string{"GET"}) {
		t.Error("Endpoint methods are not correct!")
	}

	tests := []struct {
		Source Processor
		Target Processor
	}{
		{e.Processors[0], p.GetNotifications},
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

func TestGetNotifications(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")

	type sampleSetup struct {
		Rows          sqlmock.Rows
		Notifications []NotificationModel
		Err           error
		ErrResult     *HttpError
	}

	tn := time.Now()
	tm := &NullTime{Time: tn, Valid: true}
	params := &types.NullJSONText{
		JSONText: []byte(`{"foo":"bar"}`),
		Valid:    true,
	}
	columns := []string{"issued", "read", "dismissed", "type", "parameters"}
	testCases := []sampleSetup{
		// good
		sampleSetup{
			sqlmock.NewRows(columns).
				AddRow(tn, tn, tn, "match", `{"foo":"bar"}`),
			[]NotificationModel{NotificationModel{
				Issued: tn, Read: tm, Dismissed: tm, Type: "match",
				Parameters: params,
			}},
			nil,
			nil,
		},
		// strange error
		sampleSetup{
			nil, nil, errors.New("Dummy Error"), ErrUnknown,
		},
		// no rows
		sampleSetup{
			sqlmock.NewRows(columns),
			[]NotificationModel{},
			nil,
			nil,
		},
	}

	for _, setup := range testCases {
		model := &NotificationsGetModel{}

		p := &NotificationsSelectProcessors{
			db:      db,
			handler: &URLPathHandlerWithAuth{User: 123},
		}

		if setup.Rows != nil {
			mock.ExpectQuery(
				`SELECT id, issued, read, dismissed, type, parameters ` +
					`FROM notifications ` +
					`WHERE \"user\" = (.+) AND dismissed IS NULL`,
			).
				WithArgs(p.handler.User).
				WillReturnRows(setup.Rows)
		} else if setup.Err != nil {
			mock.ExpectQuery(
				`SELECT id, issued, read, dismissed, type, parameters ` +
					`FROM notifications ` +
					`WHERE \"user\" = (.+) AND dismissed IS NULL`,
			).
				WithArgs(p.handler.User).
				WillReturnError(setup.Err)
		}

		m, err := p.GetNotifications(model)

		if err == nil && m != model {
			t.Error("Correct model was not returned.")
		}
		expected, merr := json.Marshal(setup.Notifications)
		if merr != nil {
			t.Error(merr)
		}
		actual, merr := json.Marshal(p.notifications)
		if merr != nil {
			t.Error(merr)
		}

		if string(actual) != string(expected) {
			t.Errorf("Notifications\n%s\nis not\n%s\n", actual, expected)
		}

		if err != setup.ErrResult {
			t.Error(err)
			t.Error("Email validation error was unexpected.")
		}
	}

}

func TestNotificationsGetWriter(t *testing.T) {

	tn := time.Now()
	tm := &NullTime{Time: tn, Valid: true}
	params := &types.NullJSONText{
		JSONText: types.JSONText(json.RawMessage([]byte(`{"foo":"bar"}`))),
		Valid:    true,
	}

	tests := []struct {
		Notifications []NotificationModel
		Err           *HttpError
	}{
		{
			Notifications: []NotificationModel{
				NotificationModel{1, tn, tm, tm, "match", params},
				NotificationModel{2, tn, tm, &NullTime{Valid: false},
					"verify", &types.NullJSONText{
						Valid: false,
					}},
			},
			Err: nil,
		},
	}

	for _, test := range tests {
		model := &NotificationsGetModel{}
		p := &NotificationsSelectProcessors{notifications: test.Notifications}

		out, err := p.WriteNotifications(model)

		if err != test.Err {
			t.Error(err)
		}

		expected, _ := json.Marshal(test.Notifications)

		if string(out) != string(expected) {
			t.Error("Writer returned wrong user")
		}
	}
}
