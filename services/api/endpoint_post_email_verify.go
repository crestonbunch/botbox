package api

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

// @Title Send Email verification
// @Description send a verification email for a user
// @Accept  json
// @Param   Authorization header  string    true  "Secret session token"
// @Param   email         query   string    true  "User email"
// @Success 200 plain
// @Failure 400 plain
// @Failure 500 plain
// @Resource /email
// @Router /email/verify [post]
func NewEmailVerifyPostEndpoint(a *App) *Endpoint {
	p := &EmailVerifyInsertProcessors{
		db:      a.db,
		emailer: a.emailer,
		handler: &JsonHandlerWithAuth{
			Target:  func() interface{} { return &EmailVerifyPostModel{} },
			session: &Session{db: a.db},
		},
	}

	return &Endpoint{
		Path:    "/email/verify",
		Methods: []string{"POST"},
		Handler: p.handler.HandleWithId,
		Processors: []Processor{
			p.ValidateEmail,
			p.Begin,
			p.InsertVerification,
			p.InsertNotification,
			p.SendVerification,
			p.Commit,
		},
		Writer: nil,
	}
}

type EmailVerifyInsertProcessors struct {
	db      *sqlx.DB
	tx      *sqlx.Tx
	emailer EmailerModel
	handler *JsonHandlerWithAuth
	name    string
	secret  string
}

type EmailVerifyPostModel struct {
	Email string `json:"email"`
}

func (e *EmailVerifyInsertProcessors) ValidateEmail(i interface{}) (interface{}, *HttpError) {
	model := i.(*EmailVerifyPostModel)

	if model.Email == "" {
		return nil, ErrMissingEmail
	}

	var name string
	err := e.db.Get(
		&name,
		"SELECT name FROM users WHERE email = $1",
		model.Email,
	)

	if err == sql.ErrNoRows {
		return nil, ErrEmailNotFound
	}

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	e.name = name

	return model, nil
}

func (e *EmailVerifyInsertProcessors) Begin(i interface{}) (interface{}, *HttpError) {
	tx, err := e.db.Beginx()
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}
	e.tx = tx
	return i, nil
}

func (e *EmailVerifyInsertProcessors) InsertVerification(i interface{}) (interface{}, *HttpError) {
	model := i.(*EmailVerifyPostModel)

	secret, err := Secret.Generate()

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	_, err = e.tx.Exec(
		`INSERT INTO verify_secrets (secret, email)  VALUES ($1, $2)`,
		secret, model.Email,
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	e.secret = secret

	return model, nil
}

func (e *EmailVerifyInsertProcessors) InsertNotification(i interface{}) (interface{}, *HttpError) {
	model := i.(*EmailVerifyPostModel)

	_, err := e.tx.Exec(
		`INSERT INTO notifications ("user", type)  VALUES ($1, $2)`,
		e.handler.User, NotificationTypeVerify,
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return model, nil
}

func (e *EmailVerifyInsertProcessors) SendVerification(i interface{}) (interface{}, *HttpError) {
	model := i.(*EmailVerifyPostModel)

	err := e.emailer.SendEmailVerification(model.Email, []byte(e.name), []byte(e.secret))

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return model, nil
}

func (e *EmailVerifyInsertProcessors) Commit(i interface{}) (interface{}, *HttpError) {
	err := e.tx.Commit()
	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return i, nil
}
