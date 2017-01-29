package api

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"net/http"
)

type App struct {
	db        *sqlx.DB
	recaptcha RecaptchaModel
	emailer   EmailerModel
	router    *mux.Router
}

// Build an app and return it.
func NewApp(db *sqlx.DB, recaptcha RecaptchaModel, emailer EmailerModel) *App {
	return &App{
		db:        db,
		recaptcha: recaptcha,
		emailer:   emailer,
		router:    mux.NewRouter(),
	}
}

// Attach an endpoint to the router. Attaches the app to the context of the
// request.
func (a *App) Attach(e *Endpoint) {
	a.router.HandleFunc(e.Path, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "app", a)
		e.Handle(w, r.WithContext(ctx))
	}).Methods(e.Methods...)
}

// Implements the http.Handler interface
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

type Handler func(*http.Request) (interface{}, *HttpError)
type Processor func(interface{}) (interface{}, *HttpError)
type Writer func(interface{}) ([]byte, *HttpError)

// An endpoint creates a handler for the given path with the given methods.
type Endpoint struct {
	Path    string
	Methods []string
	// The handler is the function that turns the HTTP request into a struct or
	// other data type to be passed to the processors.
	Handler Handler
	// Processors are executed in sequence, passing the output of the one into
	// the input of the next one, starting with the output returned from the
	// handler. Processors can validate input, make database calls, etc.
	Processors []Processor
	// The writer takes the output of the last processor and turns it into a byte
	// array to return to the client.
	Writer Writer
}

// Satisfies the http.Handler type to be used as a handler function for
// an HTTP server.
func (e *Endpoint) Handle(w http.ResponseWriter, r *http.Request) {
	model, err := e.Handler(r)
	if err != nil {
		http.Error(w, err.Error(), err.StatusCode())
		return
	}
	for _, p := range e.Processors {
		model, err = p(model)
		if err != nil {
			http.Error(w, err.Error(), err.StatusCode())
			return
		}
	}

	if e.Writer != nil {
		output, err := e.Writer(model)
		if err != nil {
			http.Error(w, err.Error(), err.StatusCode())
			return
		}
		fmt.Fprint(w, string(output))
	}
}
