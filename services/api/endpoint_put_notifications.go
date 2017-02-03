package api

import (
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

// @Title Put Notifications
// @Description Mark notifications as read/dismissed in a batch query.
// @Accept  json
// @Param   Authorization header  string    true  "Secret session token"
// @Param   notifications body    array     true  "Notification ids to update"
// @Param   notifications body    read      false "Mark notifications as read"
// @Param   notifications body    dismissed false "Mark notifications as dismissed"
// @Success 200 plain
// @Failure 400 plain
// @Failure 500 plain
// @Resource /notifications
// @Router /notifications [put]

// NewNotificationsPutEndpoint creates a new endpoint for updating
// notifications for an authenticated user.
func NewNotificationsPutEndpoint(a *App) *Endpoint {
	p := &NotificationsUpdateProcessors{
		db: a.db,
		handler: &JsonHandlerWithAuth{
			Target:  func() interface{} { return &NotificationsPutModel{} },
			session: &Session{db: a.db},
		},
	}

	return &Endpoint{
		Path:    "/notifications",
		Methods: []string{"PUT"},
		Handler: p.handler.HandleWithId,
		Processors: []Processor{
			p.UpdateNotifications,
		},
		Writer: nil,
	}
}

// NotificationsUpdateProcessors are a set of processors for processing the
// Get Notifications endpoint.
type NotificationsUpdateProcessors struct {
	db      *sqlx.DB
	handler *JsonHandlerWithAuth
}

// NotificationsPutModel stores information from the API request.
type NotificationsPutModel struct {
	Notifications []int `json:"notifications"`
	Read          bool  `json:"read"`
	Dismissed     bool  `json:"dismissed"`
}

// UpdateNotifications queries the database for notifications for the
// authenticated user.
func (e *NotificationsUpdateProcessors) UpdateNotifications(i interface{}) (interface{}, *HttpError) {
	model := i.(*NotificationsPutModel)
	fmt.Printf("%+v\n", model)

	if len(model.Notifications) == 0 {
		return nil, ErrMissingNotifications
	}

	set := ""
	if model.Read && !model.Dismissed {
		set = "read = NOW()"
	} else if model.Dismissed && !model.Read {
		set = "dismissed = NOW()"
	} else if model.Dismissed && model.Read {
		set = "read = NOW(), dismissed = NOW()"
	} else {
		return model, nil
	}

	params := []interface{}{e.handler.User}
	for _, n := range model.Notifications {
		params = append(params, n)
	}
	query := sqlx.Rebind(sqlx.DOLLAR,
		`UPDATE notifications
        SET `+set+`
        WHERE "user" = ? AND id IN (?`+
			strings.Repeat(",?", len(model.Notifications)-1)+`)`,
	)
	_, err := e.db.Exec(
		query,
		params...,
	)

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	return model, nil
}
