package main

import (
	"github.com/crestonbunch/botbox/services/api"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
)

type Config struct {
	DomainName       string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
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
		PostgresUser:     os.Getenv("BOTBOX_DB_USER"),
		PostgresPassword: os.Getenv("BOTBOX_DB_PASSWORD"),
		PostgresDB:       os.Getenv("BOTBOX_DB_NAME"),
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
		config.PostgresDB+" password="+config.PostgresPassword)

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

	app.Attach(api.NewUserPostEndpoint(app))
	app.Attach(api.NewEmailVerifyPostEndpoint(app))
	app.Attach(api.NewEmailVerifyPutEndpoint(app))
}
