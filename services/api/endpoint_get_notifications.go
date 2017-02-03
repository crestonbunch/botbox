package api

import (
	"encoding/json"
	"log"

	"github.com/jmoiron/sqlx"
)

// @Title Get Notifications
// @Description Get the un-dismissed notifications of a logged-in user.
// @Accept  json
// @Param   Authorization header  string   true  "Secret session token"
// @Success 200 plain
// @Failure 400 plain
// @Failure 500 plain
// @Resource /notifications
// @Router /notifications [get]

// NewNotificationsGetEndpoint creates a new endpoint for getting the list
// of notifications for an authenticated user.
func NewNotificationsGetEndpoint(a *App) *Endpoint {
	p := &NotificationsSelectProcessors{
		db: a.db,
		handler: &URLPathHandlerWithAuth{
			Target:  func() interface{} { return &NotificationsGetModel{} },
			session: &Session{db: a.db},
		},
	}

	return &Endpoint{
		Path:    "/notifications",
		Methods: []string{"GET"},
		Handler: p.handler.HandleWithId,
		Processors: []Processor{
			p.GetNotifications,
		},
		Writer: p.WriteNotifications,
	}
}

// NotificationsSelectProcessors are a set of processors for processing the
// Get Notifications endpoint.
type NotificationsSelectProcessors struct {
	db            *sqlx.DB
	handler       *URLPathHandlerWithAuth
	notifications []NotificationModel
}

// NotificationsGetModel stores information from the API request.
type NotificationsGetModel struct {
}

// GetNotifications queries the database for notifications for the authenticated
// user.
func (e *NotificationsSelectProcessors) GetNotifications(i interface{}) (interface{}, *HttpError) {
	model := i.(*NotificationsGetModel)

	rows, err := e.db.Queryx(
		`SELECT id, issued, read, dismissed, type, parameters FROM notifications
        WHERE "user" = $1 AND dismissed IS NULL`,
		e.handler.User,
	)

	if err != nil {
		log.Printf("Error getting notifications: %s\n", err)
		return nil, ErrUnknown
	}

	notifications := []NotificationModel{}
	for rows.Next() {
		n := NotificationModel{}
		err := rows.StructScan(&n)
		if err != nil {
			log.Printf("Error scanning notifications: %s\n", err)
			return nil, ErrUnknown
		}
		notifications = append(notifications, n)
	}

	e.notifications = notifications

	return model, nil
}

// WriteNotifications converts the list of notifications to JSON.
func (e *NotificationsSelectProcessors) WriteNotifications(i interface{}) ([]byte, *HttpError) {
	result, err := json.Marshal(e.notifications)

	if err != nil {
		log.Printf("Error writing notifications: %s\n", err)
		return nil, ErrUnknown
	}

	return result, nil
}
