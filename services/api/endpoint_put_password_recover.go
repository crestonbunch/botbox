package api

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"net/http"
)

// @Title Reset Password
// @Description Validate a user's email verification secret
// @Accept  json
// @Param   secret     query  string   true  "Recovery secret"
// @Param   password   query  string   true  "New password"
// @Success 200 plain
// @Failure 400 plain
// @Failure 500 plain
// @Resource /password
// @Router /password/recover [put]
func NewPasswordRecoverPutEndpoint(a *App) *Endpoint {
	p := &PasswordRecoverUpdateProcessors{
		db: a.db,
	}

	return &Endpoint{
		Path:    "/password/recover",
		Methods: []string{"PUT"},
		Handler: p.Handler,
		Processors: []Processor{
			p.ValidateSecret,
			p.ValidatePassword,
			p.Begin,
			p.UpdateSecret,
			p.UpdatePassword,
			p.Commit,
		},
		Writer: nil,
	}
}

type PasswordRecoverUpdateProcessors struct {
	db   *sqlx.DB
	tx   *sqlx.Tx
	user int
}

type PasswordRecoverPutModel struct {
	Secret   string `json:"secret"`
	Password string `json:"password"`
}

func (e *PasswordRecoverUpdateProcessors) Handler(r *http.Request) (interface{}, *HttpError) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	m := &PasswordRecoverPutModel{}
	err = json.Unmarshal(body, m)
	if err != nil {
		return nil, ErrInvalidJson
	}

	return m, nil
}

func (e *PasswordRecoverUpdateProcessors) Begin(i interface{}) (interface{}, *HttpError) {
	tx, err := e.db.Beginx()
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}
	e.tx = tx
	return i, nil
}

func (e *PasswordRecoverUpdateProcessors) ValidateSecret(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordRecoverPutModel)

	if model.Secret == "" {
		return nil, ErrInvalidSecret
	}

	var count int
	err := e.db.Get(
		&count,
		`SELECT COUNT(secret) as count FROM recovery_secrets WHERE secret = $1
		AND expires > NOW() AND used == FALSE`,
		model.Secret,
	)

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	if count == 0 {
		return nil, ErrInvalidSecret
	}

	return model, nil
}

func (e *PasswordRecoverUpdateProcessors) ValidatePassword(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordRecoverPutModel)

	if model.Password == "" {
		return nil, ErrMissingPassword
	}
	if len(model.Password) < MinPasswordLen {
		return nil, ErrPasswordTooShort
	}
	return model, nil
}

func (e *PasswordRecoverUpdateProcessors) UpdateSecret(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordRecoverPutModel)

	var user int
	err := e.tx.Get(
		&user,
		`UPDATE recovery_secrets SET used = TRUE WHERE secret = $1
		RETURNING user`, model.Secret,
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	e.user = user

	return model, nil
}

func (e *PasswordRecoverUpdateProcessors) UpdatePassword(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordRecoverPutModel)

	password, err := Hasher.hash(model.Password)
	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	_, err = e.tx.Exec(
		`UPDATE passwords SET hash = $1, salt = $2, method = $3, updated = NOW()
		WHERE user = $4`,
		password.Hash, password.Salt, password.Method, e.user,
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return nil, nil
}

func (e *PasswordRecoverUpdateProcessors) Commit(i interface{}) (interface{}, *HttpError) {
	err := e.tx.Commit()
	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return i, nil
}
