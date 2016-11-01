package main

import (
	"database/sql"
	"github.com/crestonbunch/botbox/services/api"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"net/smtp"
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

	r := mux.NewRouter()

	// These are endpoints for account management. They handle things such as
	// registering a new user, updating passwords, etc.
	accountNew := api.RequestWrapper(api.AccountNew(sendValidation), db)
	accountVerify := api.RequestWrapper(api.AccountVerify, db)
	accountChangePw := api.RequestWrapper(api.AccountChangePassword, db)
	//accountForgotPw := api.RequestWrapper(api.AccountForgotPassword(sendRecovery), db)
	r.HandleFunc("/account/user/new", accountNew).Methods("POST")
	r.HandleFunc("/account/user/verify", accountVerify).Methods("POST")
	r.HandleFunc("/account/password/change", accountChangePw).Methods("POST")
	//r.HandleFunc("/account/password/recover", accountForgotPw).Methods("POST")

	// These are endpoints for user-related requests. They handle such things as
	// getting information about a certain user, etc.
	userGet := api.RequestWrapper(api.UserGet, db)
	r.HandleFunc("/user/{username}", userGet).Methods("GET")

	// These are endpoints for session management. They handle things such as
	// getting the current user, authorizing a new session, renewing, revoking,
	// etc. Some API calls will not work unless a session is first authorized to
	// make them (i.e., by logging in).
	/*
		sessionUser := api.RequestWrapper(api.SessionUser, db)
		sessionAuth := api.RequestWrapper(api.SessionAuth, db)
		sessionRenew := api.RequestWrapper(api.SessionRenew, db)
		sessionRevoke := api.RequestWrapper(api.SessionRevoke, db)
		r.HandleFunc("/session/user", sessionUser).Methods("GET")
		r.HandleFunc("/session/authorize", sessionAuth).Methods("POST")
		r.HandleFunc("/session/renew", sessionRenew).Methods("POST")
		r.HandleFunc("/session/revoke", sessionRevoke).Methods("POST")
	*/

	log.Fatal(http.ListenAndServe(":8081", r))
}

func sendValidation(db *sql.DB, username, email string) error {
	domain := os.Getenv(EnvVarDomainName)
	smtpId := os.Getenv(EnvVarSmtpIdentity)
	smtpUser := os.Getenv(EnvVarSmtpUsername)
	smtpPass := os.Getenv(EnvVarSmtpPassword)
	smtpHost := os.Getenv(EnvVarSmtpHost)
	smtpPort := os.Getenv(EnvVarSmtpPort)
	smtpAuth := smtp.PlainAuth(smtpId, smtpUser, smtpPass, smtpHost)
	smtpAddr := smtpHost + ":" + smtpPort

	verifySecret, err := api.CreateVerification(db, username, email)
	if err != nil {
		return err
	}
	verifyUrl := "http://" + domain + "/verify/" + verifySecret

	to := []string{email}
	msg := []byte(
		"To: " + email + "\r\n" +
			"Subject: " + "Email verification \r\n" +
			"\r\n" +
			"Please use the following link to verify your account:\r\n" +
			verifyUrl,
	)
	err = smtp.SendMail(smtpAddr, smtpAuth, "no-reply@"+domain, to, msg)

	return err
}
