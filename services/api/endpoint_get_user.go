package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"

	"github.com/jmoiron/sqlx"
)

// @Title Get user
// @Description Get a user by id
// @Accept  json
// @Param   id      path     string     true     "User id"
// @Success 200 plain
// @Failure 404 plain
// @Failure 500 plain
// @Resource /user
// @Router /user/id/{id} [get]

// NewUserIdGetEndpoint creates a new endpoint for getting users by id.
func NewUserIdGetEndpoint(a *App) *Endpoint {
	p := &UserIdGetProcessors{
		db: a.db,
		handler: &URLPathHandler{
			Target: func() interface{} { return &UserIdSelectModel{} },
		},
	}

	return &Endpoint{
		Path:    "/user/id/{id}",
		Methods: []string{"GET"},
		Handler: p.handler.Handle,
		Processors: []Processor{
			p.GetUser,
			p.GetPermissions,
		},
		Writer: p.UserWriter,
	}
}

// UserIdSelectModel holds the data from the user request.
type UserIdSelectModel struct {
	Id string `json:'id'`
}

// UserIdGetProcessors holds data about the processors handling the request.
type UserIdGetProcessors struct {
	db      *sqlx.DB
	user    *User
	handler *URLPathHandler
}

// GetUser gets the user from the database.
func (e *UserIdGetProcessors) GetUser(i interface{}) (interface{}, *HttpError) {
	model := i.(*UserIdSelectModel)

	if model.Id == "" {
		return nil, ErrMissingParameter
	}

	id, err := strconv.Atoi(model.Id)

	if err != nil {
		return nil, ErrNotAnInteger
	}

	var user User
	err = e.db.Get(
		&user,
		`SELECT id, name, joined, permission_set FROM users WHERE id = $1`,
		id,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	e.user = &user

	return model, nil
}

// GetPermissions queries the user permissions from the database.
func (e *UserIdGetProcessors) GetPermissions(i interface{}) (interface{}, *HttpError) {
	model := i.(*UserIdSelectModel)

	rows, err := e.db.Queryx(
		`SELECT permission FROM permission_set_permissions
        WHERE permission_set = $1`,
		e.user.PermissionSet,
	)

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			log.Println(err)
			return nil, ErrUnknown
		}
		e.user.Permissions = append(e.user.Permissions, perm)
	}

	return model, nil
}

// UserWriter returns the JSON output
func (e *UserIdGetProcessors) UserWriter(i interface{}) ([]byte, *HttpError) {
	s, err := json.Marshal(e.user)
	if err != nil {
		log.Println(err)
		return []byte{}, ErrUnknown
	}

	return s, nil
}
