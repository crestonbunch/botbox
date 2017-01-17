package api

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"net/http"
)

// @Title Change Password
// @Description Change the password of a logged-in user.
// @Accept  json
// @Param   Authorization header  string   true  "Secret session token"
// @Param   old           query   string   true  "Old password"
// @Param   new           query   string   true  "New password"
// @Success 200 plain
// @Failure 400 plain
// @Failure 500 plain
// @Resource /password
// @Router /password [put]
func NewPasswordPutEndpoint(a *App) *Endpoint {
	p := &PasswordUpdateProcessors{
		db: a.db,
	}

	return &Endpoint{
		Path:       "/password",
		Methods:    []string{"PUT"},
		Handler:    p.Handler,
		Processors: []Processor{},
		Writer:     nil,
	}
}

type PasswordUpdateProcessors struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

type PasswordPutModel struct {
	Session string
	Old     string `json:"old"`
	New     string `json:"new"`
}

func (e *PasswordUpdateProcessors) Handler(r *http.Request) (interface{}, *HttpError) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	m := &PasswordPutModel{}
	err = json.Unmarshal(body, m)
	if err != nil {
		return nil, ErrInvalidJson
	}

	m.Session = r.Header.Get("Authorization")

	return m, nil
}

func (e *PasswordUpdateProcessors) ValidateSecret(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordPutModel)

	if model.Session == "" {
		return nil, ErrInvalidSecret
	}

	var response struct {
		Count int `db:"count"`
		User  int `db:"user"`
	}
	err := e.db.Get(
		&response,
		`SELECT COUNT(secret) as count, user FROM session_secrets WHERE secret = $1
		AND expires > NOW() AND revoked == FALSE`,
		model.Session,
	)

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	if response.Count == 0 {
		return nil, ErrInvalidSecret
	}

	return model, nil
}
