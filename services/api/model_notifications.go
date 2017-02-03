package api

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

// NotificationTypeVerify is type of notification for email verification.
const NotificationTypeVerify = "verify"

// NotificationModel holds information about a notification from the database.
type NotificationModel struct {
	ID         int                 `json:"id" db:"id"`
	Issued     time.Time           `json:"issued" db:"issued"`
	Read       *NullTime           `json:"read" db:"read"`
	Dismissed  *NullTime           `json:"dismissed" db:"dismissed"`
	Type       string              `json:"type" db:"type"`
	Parameters *types.NullJSONText `json:"parameters" db:"parameters"`
}
