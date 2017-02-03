package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/crestonbunch/botbox/services/api"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DomainName       string
	PostgresHost     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string
	SMTPHost         string
	SMTPPort         string
	SMTPName         string
	SMTPPassword     string
	RecaptchaKey     string
	RecaptchaSecret  string
}

func ConfigFromEnv() *Config {
	return &Config{
		DomainName:       os.Getenv("BOTBOX_DOMAIN_NAME"),
		PostgresHost:     os.Getenv("BOTBOX_DB_HOST"),
		PostgresUser:     os.Getenv("BOTBOX_DB_USER"),
		PostgresPassword: os.Getenv("BOTBOX_DB_PASSWORD"),
		PostgresDB:       os.Getenv("BOTBOX_DB_NAME"),
		PostgresSSLMode:  os.Getenv("BOTBOX_DB_SSLMODE"),
		SMTPHost:         os.Getenv("BOTBOX_SMTP_HOST"),
		SMTPPort:         os.Getenv("BOTBOX_SMTP_PORT"),
		SMTPName:         os.Getenv("BOTBOX_SMTP_USERNAME"),
		SMTPPassword:     os.Getenv("BOTBOX_SMTP_PASSWORD"),
		RecaptchaKey:     os.Getenv("BOTBOX_RECAPTCHA_SITEKEY"),
		RecaptchaSecret:  os.Getenv("BOTBOX_RECAPTCHA_SECRET"),
	}
}

type Templates struct {
	EmailVerification []byte
}

func TemplatesFromFiles() (*Templates, error) {
	templates := &Templates{}

	if f, err := os.OpenFile("emails/verify.html", os.O_RDONLY, os.ModePerm); err == nil {
		if templates.EmailVerification, err = ioutil.ReadAll(f); err != nil {
			return nil, err
		}
		f.Close()
	} else {
		return nil, err
	}

	return templates, nil
}

func main() {
	config := ConfigFromEnv()
	db := sqlx.MustConnect("postgres", "user="+config.PostgresUser+" dbname="+
		config.PostgresDB+" password="+config.PostgresPassword+
		" host="+config.PostgresHost+" sslmode="+config.PostgresSSLMode)

	recaptcha := &api.Recaptcha{
		RecaptchaUrl:    api.RecaptchaUrl,
		RecaptchaSecret: config.RecaptchaSecret,
	}

	templates, err := TemplatesFromFiles()

	if err != nil {
		log.Fatal(err)
	}

	emailer := &api.Emailer{
		Auth: smtp.PlainAuth(
			"", config.SMTPName, config.SMTPPassword, config.SMTPHost,
		),
		Server: config.SMTPHost + ":" + config.SMTPPort,
		Domain: config.DomainName,
		EmailVerificationTemplate: templates.EmailVerification,
	}

	app := api.NewApp(db, recaptcha, emailer)

	app.Attach(api.NewUserIdGetEndpoint(app))
	app.Attach(api.NewUserPostEndpoint(app))

	app.Attach(api.NewEmailGetEndpoint(app))
	app.Attach(api.NewEmailVerifyPostEndpoint(app))
	app.Attach(api.NewEmailVerifyPutEndpoint(app))

	app.Attach(api.NewPasswordPutEndpoint(app))
	app.Attach(api.NewPasswordRecoverPostEndpoint(app))
	app.Attach(api.NewPasswordRecoverPutEndpoint(app))

	app.Attach(api.NewSessionPostEndpoint(app))
	app.Attach(api.NewSessionPutEndpoint(app))

	app.Attach(api.NewNotificationsGetEndpoint(app))
	app.Attach(api.NewNotificationsPutEndpoint(app))

	srv := &http.Server{
		Handler: app,
		Addr:    ":8081",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Serving...")
	log.Fatal(srv.ListenAndServe())
}
