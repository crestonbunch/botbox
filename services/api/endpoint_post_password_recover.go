package api

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"net/http"
)

// @Title Send Password Recovery
// @Description send a password recovery email for a user
// @Accept  json
// @Param   email     query  string   true  "User email"
// @Success 200 plain
// @Failure 400 plain
// @Failure 500 plain
// @Resource /password
// @Router /password/recover [post]
func NewPasswordRecoverPostEndpoint(a *App) *Endpoint {
	p := &PasswordRecoverInsertProcessers{
		db:      a.db,
		emailer: a.emailer,
	}

	return &Endpoint{
		Path:    "/password/recover",
		Methods: []string{"POST"},
		Handler: p.Handler,
		Processors: []Processor{
			p.ValidateEmail,
			p.Begin,
			p.InsertRecovery,
			p.SendRecovery,
			p.Commit,
		},
		Writer: nil,
	}
}

type PasswordRecoverInsertProcessers struct {
	db      *sqlx.DB
	tx      *sqlx.Tx
	emailer EmailerModel
	user    int
	name    string
	secret  string
}

type PasswordRecoverPostModel struct {
	Email string `json:"email"`
}

func (e *PasswordRecoverInsertProcessers) Handler(r *http.Request) (interface{}, *HttpError) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	m := &PasswordRecoverPostModel{}
	err = json.Unmarshal(body, m)
	if err != nil {
		return nil, ErrInvalidJson
	}

	return m, nil
}

func (e *PasswordRecoverInsertProcessers) ValidateEmail(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordRecoverPostModel)

	if model.Email == "" {
		return nil, ErrMissingEmail
	}

	var result struct {
		Count int    `db:"count"`
		Name  string `db:"name"`
		User  int    `db:"id"`
	}
	err := e.db.Get(
		&result,
		"SELECT COUNT(id) as count, name, id FROM users WHERE email = $1",
		model.Email,
	)

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	if result.Count == 0 {
		return nil, ErrEmailNotFound
	}

	e.name = result.Name
	e.user = result.User

	return model, nil
}

func (e *PasswordRecoverInsertProcessers) Begin(i interface{}) (interface{}, *HttpError) {
	tx, err := e.db.Beginx()
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}
	e.tx = tx
	return i, nil
}

func (e *PasswordRecoverInsertProcessers) InsertRecovery(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordRecoverPostModel)

	secret, err := Secret.Generate()

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	_, err = e.tx.Exec(
		`INSERT INTO recovery_secrets (secret, user)  VALUES ($1, $2)`,
		secret, e.user,
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	e.secret = secret

	return model, nil
}

func (e *PasswordRecoverInsertProcessers) SendRecovery(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordRecoverPostModel)

	err := e.emailer.SendPasswordRecovery(model.Email, []byte(e.name), []byte(e.secret))

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return model, nil
}

func (e *PasswordRecoverInsertProcessers) Commit(i interface{}) (interface{}, *HttpError) {
	err := e.tx.Commit()
	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return i, nil
}
