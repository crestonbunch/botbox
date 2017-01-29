package api

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

// @Title Renew session token
// @Description renew a session token for authenticating user API calls
// @Accept  json
// @Param   secret     query  string   true  "Secret session token"
// @Success 200 plain string "The new session token for authenticating"
// @Failure 400 plain
// @Failure 500 plain
// @Resource /session
// @Router /session [put]
func NewSessionPutEndpoint(a *App) *Endpoint {
	p := &SessionUpdateProcessors{
		db: a.db,
		handler: &JsonHandler{
			Target: func() interface{} { return &SessionPutModel{} },
		},
	}

	return &Endpoint{
		Path:    "/session",
		Methods: []string{"PUT"},
		Handler: p.handler.Handle,
		Processors: []Processor{
			p.ValidateSecret,
			p.CreateSecret,
		},
		Writer: p.Write,
	}
}

type SessionUpdateProcessors struct {
	db      *sqlx.DB
	handler *JsonHandler
	user    int
	secret  string
}

type SessionPutModel struct {
	Secret string `json:"secret"`
}

func (e *SessionUpdateProcessors) ValidateSecret(i interface{}) (interface{}, *HttpError) {
	model := i.(*SessionPutModel)

	if model.Secret == "" {
		return nil, ErrInvalidSecret
	}

	var user int
	err := e.db.Get(
		&user,
		`SELECT "user" FROM session_secrets WHERE secret = $1 AND expires < NOW()
		AND revoked == FALSE`,
		model.Secret,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidSecret
	}

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	e.user = user

	return model, nil
}

func (e *SessionUpdateProcessors) CreateSecret(i interface{}) (interface{}, *HttpError) {
	model := i.(*SessionPutModel)

	secret, err := Secret.Generate()

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	_, err = e.db.Exec(
		`INSERT INTO session_secrets (secret, "user")  VALUES ($1, $2)`,
		secret, e.user,
	)

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	e.secret = secret

	return model, nil
}

func (e *SessionUpdateProcessors) Write(i interface{}) ([]byte, *HttpError) {
	return []byte(e.secret), nil
}
