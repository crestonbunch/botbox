package api

import (
	"github.com/jmoiron/sqlx"
	"log"
)

// @Title Verify Email
// @Description Validate a user's email verification secret
// @Accept  json
// @Param   secret     query  string   true  "Verification secret"
// @Success 200 plain
// @Failure 400 plain
// @Failure 500 plain
// @Resource /email
// @Router /email/verify [put]
func NewEmailVerifyPutEndpoint(a *App) *Endpoint {
	p := &EmailVerifyUpdateProcessors{
		db: a.db,
		handler: &JsonHandler{
			Target: func() interface{} { return &EmailVerifyPutModel{} },
		},
	}

	return &Endpoint{
		Path:    "/email/verify",
		Methods: []string{"PUT"},
		Handler: p.handler.Handle,
		Processors: []Processor{
			p.ValidateSecret,
			p.Begin,
			p.UpdateSecret,
			p.UpdateUser,
			p.Commit,
		},
		Writer: nil,
	}
}

type EmailVerifyUpdateProcessors struct {
	db      *sqlx.DB
	tx      *sqlx.Tx
	handler *JsonHandler
	email   string
}

type EmailVerifyPutModel struct {
	Secret string `json:"secret"`
}

func (e *EmailVerifyUpdateProcessors) Begin(i interface{}) (interface{}, *HttpError) {
	tx, err := e.db.Beginx()
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}
	e.tx = tx
	return i, nil
}

func (e *EmailVerifyUpdateProcessors) ValidateSecret(i interface{}) (interface{}, *HttpError) {
	model := i.(*EmailVerifyPutModel)

	if model.Secret == "" {
		return nil, ErrInvalidSecret
	}

	var count int
	err := e.db.Get(
		&count,
		`SELECT COUNT(secret) as count FROM verify_secrets WHERE secret = $1 AND 
		expires > NOW() AND used == FALSE`,
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

func (e *EmailVerifyUpdateProcessors) UpdateSecret(i interface{}) (interface{}, *HttpError) {
	model := i.(*EmailVerifyPutModel)

	var email string
	err := e.tx.Get(
		&email,
		`UPDATE verify_secrets SET used = TRUE WHERE secret = $1
		RETURNING email`, model.Secret,
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	e.email = email

	return model, nil
}

func (e *EmailVerifyUpdateProcessors) UpdateUser(i interface{}) (interface{}, *HttpError) {
	_, err := e.tx.Exec(
		`UPDATE users SET permission_set = $1 WHERE email = $2 AND
		permission_set = $3`,
		"VERIFIED", e.email, "UNVERIFIED",
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return nil, nil
}

func (e *EmailVerifyUpdateProcessors) Commit(i interface{}) (interface{}, *HttpError) {
	err := e.tx.Commit()
	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return i, nil
}
