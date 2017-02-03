package api

import (
	"github.com/jmoiron/sqlx"
)

// A generic interface for something that retrieves session data.
type SessionModel interface {
	GetUserId(string) (int, error)
	GetPermissions(string) (PermissionSet, error)
}

// A concrete implementation of SessionModel for a Postgres backend.
type Session struct {
	db *sqlx.DB
}

func (s *Session) GetUserId(secret string) (int, error) {

	var id int
	err := s.db.Get(
		&id,
		`SELECT id FROM users INNER JOIN session_secrets 
		ON (users.id = session_secrets.user)
		WHERE secret = $1 AND expires > NOW() AND revoked = FALSE`,
		secret,
	)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Session) GetPermissions(secret string) (PermissionSet, error) {

	rows, err := s.db.Queryx(
		`SELECT permission FROM permission_set_permissions
		INNER JOIN users 
		 ON (users.permission_set = permission_set_permissions.permission_set)
		INNER JOIN session_secrets 
		 ON (session_secrets.user = users.id)
		WHERE session_secrets.secret = $1 
		AND users.id = session_secrets.user
		AND permission_set_permissions.permission_set = users.permission_set`,
		secret,
	)

	if err != nil {
		return nil, err
	}

	permissions := PermissionSet([]string{})

	for rows.Next() {
		var result string
		err = rows.Scan(&result)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, result)
	}

	return permissions, nil
}
