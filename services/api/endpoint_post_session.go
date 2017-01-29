package api

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

// @Title Create session token
// @Description create a session token for authenticating user API calls
// @Accept  json
// @Param   email     query  string   true  "User email"
// @Param   password  query  string   true  "User password"
// @Success 200 plain string "The session token for authenticating"
// @Failure 400 plain
// @Failure 500 plain
// @Resource /session
// @Router /session [post]
func NewSessionPostEndpoint(a *App) *Endpoint {
	p := &SessionInsertProcessors{
		db: a.db,
		handler: &JsonHandler{
			Target: func() interface{} { return &SessionPostModel{} },
		},
	}

	return &Endpoint{
		Path:    "/session",
		Methods: []string{"POST"},
		Handler: p.handler.Handle,
		Processors: []Processor{
			p.ValidateEmail,
			p.ValidatePassword,
			p.CreateSecret,
		},
		Writer: p.Write,
	}
}

type SessionInsertProcessors struct {
	db      *sqlx.DB
	handler *JsonHandler
	user    int
	secret  string
}

type SessionPostModel struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (e *SessionInsertProcessors) ValidateEmail(i interface{}) (interface{}, *HttpError) {
	model := i.(*SessionPostModel)

	if model.Email == "" {
		return nil, ErrMissingEmail
	}

	var user int
	err := e.db.Get(
		&user,
		"SELECT id FROM users WHERE email = $1",
		model.Email,
	)

	if err == sql.ErrNoRows {
		return nil, ErrLoginIncorrect
	}

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	e.user = user

	return model, nil
}

func (e *SessionInsertProcessors) ValidatePassword(i interface{}) (interface{}, *HttpError) {
	model := i.(*SessionPostModel)

	pass := &Password{}
	e.db.Get(pass, `SELECT "user", method, hash, salt FROM passwords
		WHERE "user" = $1`, e.user)

	if pass.Matches(model.Password) != nil {
		return nil, ErrLoginIncorrect
	}

	return model, nil
}

func (e *SessionInsertProcessors) CreateSecret(i interface{}) (interface{}, *HttpError) {
	model := i.(*SessionPostModel)

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

func (e *SessionInsertProcessors) Write(i interface{}) ([]byte, *HttpError) {
	return []byte(e.secret), nil
}
