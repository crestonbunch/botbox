package api

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// NullTime is a nullable database type that correctly serializes into JSON.
type NullTime pq.NullTime

// MarshalJSON converts a NullTime into null if it is null, or a marshalled
// time.Time if not null.
func (t *NullTime) MarshalJSON() ([]byte, error) {
	if t.Valid {
		return t.Time.MarshalJSON()
	}

	return json.Marshal(nil)
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}
