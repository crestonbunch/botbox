package api

import (
	"log"

	"github.com/jmoiron/sqlx"
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
		handler: &JsonHandlerWithAuth{
			Target:  func() interface{} { return &PasswordPutModel{} },
			session: &Session{db: a.db},
		},
	}

	return &Endpoint{
		Path:    "/password",
		Methods: []string{"PUT"},
		Handler: p.handler.HandleWithId,
		Processors: []Processor{
			p.CheckPasswordMatch,
			p.ValidatePassword,
			p.UpdatePassword,
		},
		Writer: nil,
	}
}

type PasswordUpdateProcessors struct {
	db      *sqlx.DB
	handler *JsonHandlerWithAuth
}

type PasswordPutModel struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func (e *PasswordUpdateProcessors) CheckPasswordMatch(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordPutModel)

	pass := &Password{}
	e.db.Get(pass, `SELECT user, method, hash, salt FROM passwords
		WHERE "user" = $1`, e.handler.User)

	if pass.Matches(model.Old) != nil {
		return nil, ErrInvalidPassword
	}

	return model, nil
}

func (e *PasswordUpdateProcessors) ValidatePassword(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordPutModel)

	return model, ValidatePassword(model.New)
}

func (e *PasswordUpdateProcessors) UpdatePassword(i interface{}) (interface{}, *HttpError) {
	model := i.(*PasswordPutModel)

	password, err := Hasher.hash(model.New)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	_, err = e.db.Exec(
		`UPDATE passwords (salt, hash, method) VALUES ($1, $2, $3) 
		WHERE "user" = $4`,
		password.Salt, password.Hash, password.Method, e.handler.User,
	)

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	return model, nil
}
