package main

import (
	"database/sql"
	"encoding/json"
	"github.com/crestonbunch/botbox/services/api"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
)

const (
	EnvVarPostgresUser     = "POSTGRES_DB_USER"
	EnvVarPostgresDb       = "POSTGRES_DB_NAME"
	EnvVarPostgresPassword = "POSTGRES_DB_PASSWORD"
	EnvVarPostgresHost     = "POSTGRES_DB_HOST"

	EnvVarSmtpIdentity = "SMTP_IDENTITY"
	EnvVarSmtpUsername = "SMTP_USERNAME"
	EnvVarSmtpPassword = "SMTP_PASSWORD"
	EnvVarSmtpHost     = "SMTP_HOST"
	EnvVarSmtpPort     = "SMTP_PORT"

	EnvVarRecaptchaSecret = "RECAPTCHA_SECRET"

	EnvVarDomainName = "BOTBOX_DOMAIN_NAME"
)

func main() {

	dbUser := os.Getenv(EnvVarPostgresUser)
	dbName := os.Getenv(EnvVarPostgresDb)
	dbPass := os.Getenv(EnvVarPostgresPassword)
	dbHost := os.Getenv(EnvVarPostgresHost)
	connStr := "postgres://" + dbUser + ":" + dbPass + "@" + dbHost +
		"/" + dbName + "?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	userModel := api.NewUserModel(db)
	emailSender := &emailSender{}
	botChecker := &botChecker{}

	accountEndpoints := api.NewAccountEndpoints(
		userModel, emailSender, botChecker,
	)
	userEndpoints := api.NewUserEndpoints(userModel)

	r := mux.NewRouter()
	userEndpoints.Attach(r)
	accountEndpoints.Attach(r)

	log.Fatal(http.ListenAndServe(":8081", r))
}

type emailSender struct{}

func (s *emailSender) SendVerificationEmail(email, secret string) error {
	domain := os.Getenv(EnvVarDomainName)
	smtpId := os.Getenv(EnvVarSmtpIdentity)
	smtpUser := os.Getenv(EnvVarSmtpUsername)
	smtpPass := os.Getenv(EnvVarSmtpPassword)
	smtpHost := os.Getenv(EnvVarSmtpHost)
	smtpPort := os.Getenv(EnvVarSmtpPort)
	smtpAuth := smtp.PlainAuth(smtpId, smtpUser, smtpPass, smtpHost)
	smtpAddr := smtpHost + ":" + smtpPort

	verifyUrl := "http://" + domain + "/verify/" + secret

	to := []string{email}
	msg := []byte(
		"To: " + email + "\r\n" +
			"Subject: " + "Email verification \r\n" +
			"\r\n" +
			"Please use the following link to verify your account:\r\n" +
			verifyUrl,
	)
	err := smtp.SendMail(smtpAddr, smtpAuth, "no-reply@"+domain, to, msg)

	return err
}

func (s *emailSender) SendPasswordRecoveryEmail(email, secret string) error {
	domain := os.Getenv(EnvVarDomainName)
	smtpId := os.Getenv(EnvVarSmtpIdentity)
	smtpUser := os.Getenv(EnvVarSmtpUsername)
	smtpPass := os.Getenv(EnvVarSmtpPassword)
	smtpHost := os.Getenv(EnvVarSmtpHost)
	smtpPort := os.Getenv(EnvVarSmtpPort)
	smtpAuth := smtp.PlainAuth(smtpId, smtpUser, smtpPass, smtpHost)
	smtpAddr := smtpHost + ":" + smtpPort

	recoverUrl := "http://" + domain + "/recover/" + secret

	to := []string{email}
	msg := []byte(
		"To: " + email + "\r\n" +
			"Subject: " + "Password recovery \r\n" +
			"\r\n" +
			"Please use the following link to recover your password:\r\n" +
			recoverUrl + "\r\n" +
			"\r\n If you did not make this request, then the link will expire" +
			"in 7 days.",
	)
	err := smtp.SendMail(smtpAddr, smtpAuth, "no-reply@"+domain, to, msg)

	return err
}

type botChecker struct{}

func (c *botChecker) IsHuman(token string) (bool, error) {
	vals := url.Values{}
	vals.Set("secret", os.Getenv(EnvVarRecaptchaSecret))
	vals.Set("response", token)

	response, err := http.PostForm(
		"https://www.google.com/recaptcha/api/siteverify",
		vals,
	)

	if err != nil {
		log.Println(err)
		return false, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return false, err
	}

	msg := struct {
		Success   bool   `json:'success'`
		Timestamp string `json:'challenge_ts'`
		Hostname  string `json:'hostname'`
	}{}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return msg.Success, nil
}
