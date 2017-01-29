package api

import (
	"log"

	"github.com/jmoiron/sqlx"
)

const (
	MaxNameLen = 20
)

// @Title insertUserWithPassword
// @Description register a user with a password
// @Accept  json
// @Param   name       query     string     true     "User display name"
// @Param   password   query     string     true     "User password"
// @Param   email      query     string     true     "User email"
// @Param   captcha    query     string     true     "User reCaptcha response"
// @Success 200 plain
// @Failure 400 plain
// @Failure 500 plain
// @Resource /user
// @Router /user [post]
func NewUserPostEndpoint(a *App) *Endpoint {
	p := &UserInsertProcessors{
		db:        a.db,
		recaptcha: a.recaptcha,
		handler: &JsonHandler{
			Target: func() interface{} { return &UserPasswordPostModel{} },
		},
	}

	return &Endpoint{
		Path:    "/user",
		Methods: []string{"POST"},
		Handler: p.handler.Handle,
		Processors: []Processor{
			p.ValidateName,
			p.ValidateEmail,
			p.ValidatePassword,
			p.ValidateCaptcha,
			p.Begin,
			p.InsertUser,
			p.InsertPassword,
			p.Commit,
		},
		Writer: nil,
	}
}

type UserPasswordPostModel struct {
	Name     string `json:'name'`
	Password string `json:'password'`
	Email    string `json:'email'`
	Captcha  string `json:'captcha'`
}

type UserInsertProcessors struct {
	db        *sqlx.DB
	tx        *sqlx.Tx
	handler   *JsonHandler
	userId    int
	recaptcha RecaptchaModel
}

func (e *UserInsertProcessors) ValidateName(i interface{}) (interface{}, *HttpError) {
	model := i.(*UserPasswordPostModel)

	if model.Name == "" {
		return nil, ErrMissingName
	}
	if len(model.Name) > MaxNameLen {
		return nil, ErrNameTooLong
	}
	return model, nil
}

func (e *UserInsertProcessors) ValidateEmail(i interface{}) (interface{}, *HttpError) {
	model := i.(*UserPasswordPostModel)

	if model.Email == "" {
		return nil, ErrMissingEmail
	}

	var count int
	err := e.db.Get(
		&count,
		"SELECT COUNT(id) as count FROM users WHERE email = $1",
		model.Email,
	)

	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	if count > 0 {
		return nil, ErrEmailInUse
	}

	return model, nil
}

func (e *UserInsertProcessors) ValidatePassword(i interface{}) (interface{}, *HttpError) {
	model := i.(*UserPasswordPostModel)

	return model, ValidatePassword(model.Password)
}

func (e *UserInsertProcessors) ValidateCaptcha(i interface{}) (interface{}, *HttpError) {
	model := i.(*UserPasswordPostModel)

	human, err := e.recaptcha.Verify(model.Captcha)
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}

	if !human {
		return nil, ErrBotDetected
	}

	return model, nil
}

func (e *UserInsertProcessors) Begin(i interface{}) (interface{}, *HttpError) {
	tx, err := e.db.Beginx()
	if err != nil {
		log.Println(err)
		return nil, ErrUnknown
	}
	e.tx = tx
	return i, nil
}

func (e *UserInsertProcessors) InsertUser(i interface{}) (interface{}, *HttpError) {
	model := i.(*UserPasswordPostModel)

	var id int
	err := e.tx.Get(&id,
		`INSERT INTO users (name, email, permission_set) VALUES ($1, $2, $3)
		RETURNING id`,
		model.Name, model.Email, "UNVERIFIED",
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}
	e.userId = id

	return model, nil
}

func (e *UserInsertProcessors) InsertPassword(i interface{}) (interface{}, *HttpError) {
	model := i.(*UserPasswordPostModel)

	password, err := Hasher.hash(model.Password)
	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	_, err = e.tx.Exec(
		`INSERT INTO passwords ("user", hash, salt, method) VALUES ($1, $2, $3, $4)`,
		e.userId, password.Hash, password.Salt, password.Method,
	)

	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return model, nil
}

func (e *UserInsertProcessors) Commit(i interface{}) (interface{}, *HttpError) {
	err := e.tx.Commit()
	if err != nil {
		log.Println(err)
		e.tx.Rollback()
		return nil, ErrUnknown
	}

	return i, nil
}
