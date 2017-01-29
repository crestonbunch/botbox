package api

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// @Title Get email
// @Description Check in an email is in use.
// @Accept  json
// @Param   email      path     string     true     "User email"
// @Success 200 plain
// @Failure 404 plain
// @Failure 500 plain
// @Resource /email
// @Router /email/{email} [get]
func NewEmailGetEndpoint(a *App) *Endpoint {
	p := &EmailSelectProcessors{
		db: a.db,
		handler: &URLPathHandler{
			Target: func() interface{} { return &EmailGetModel{} },
		},
	}

	return &Endpoint{
		Path:    "/email/{email}",
		Methods: []string{"GET"},
		Handler: p.handler.Handle,
		Processors: []Processor{
			p.ValidateEmail,
		},
		Writer: nil,
	}
}

type EmailGetModel struct {
	Email string `json:'email'`
}

type EmailSelectProcessors struct {
	db      *sqlx.DB
	handler *URLPathHandler
}

func (e *EmailSelectProcessors) ValidateEmail(i interface{}) (interface{}, *HttpError) {
	model := i.(*EmailGetModel)

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

	if count == 0 {
		return nil, ErrEmailNotFound
	}

	return model, nil
}
