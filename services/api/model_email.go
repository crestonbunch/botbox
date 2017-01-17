package api

import (
	"bytes"
	"net/smtp"
)

const EmailSender = "mailbot"
const EmailSenderName = "Botbox"

var EmailSecretToken = []byte("{{secret}}")
var EmailNameToken = []byte("{{name}}")
var EmailDomainToken = []byte("{{domain}}")

// A generic interface for an emailer that sends emails to clients.
type EmailerModel interface {
	// Send an email with a secret key to verify that the email exists and is
	// valid.
	SendEmailVerification(email string, name, secret []byte) error
	// Send a password recovery email to reset an account's password.
	SendPasswordRecovery(email string, name, secret []byte) error
}

// A global implementation of EmailerModel which sends emails via SMTP.
type Emailer struct {
	Auth smtp.Auth

	// The SMTP server to use (with port)
	Server string

	// The service's domain name. E.g., 'example.com'
	Domain string

	// EmailVerificationTemplate may contain the following substrings:
	// '{{secret}}' -- replaced with the verification secret
	// '{{name}}' -- replaced with the user's name
	// '{{domain}}' -- replaced with the domain name
	EmailVerificationTemplate []byte

	// PasswordRecoveryTemplate may contain the following substrings:
	// '{{secret}}' -- replaced with the verification secret
	// '{{name}}' -- replaced with the user's name
	// '{{domain}}' -- replaced with the domain name
	PasswordRecoveryTemplate []byte
}

func (e *Emailer) SendEmailVerification(email string, name, secret []byte) error {
	from := EmailSender + "@" + e.Domain
	msg := e.BuildEmailVerification(email, name, secret)
	return e.SendEmail(e.Server, from, email, e.Auth, msg)
}

func (e *Emailer) BuildEmailVerification(email string, name, secret []byte) []byte {
	template := e.EmailVerificationTemplate
	template = bytes.Replace(template, EmailSecretToken, secret, -1)
	template = bytes.Replace(template, EmailNameToken, name, -1)
	template = bytes.Replace(template, EmailDomainToken, []byte(e.Domain), -1)

	from := EmailSender + "@" + e.Domain
	msg := bytes.NewBufferString("To: ")
	msg.Write(name)
	msg.WriteString(" <" + email + ">\r\n")
	msg.WriteString("From: ")
	msg.WriteString(EmailSenderName)
	msg.WriteString(" <" + from + ">\r\n")
	msg.WriteString("Subject: Email Verification\r\n")
	msg.WriteString("\r\n")
	msg.Write(template)
	msg.WriteString("\r\n")

	return msg.Bytes()
}

func (e *Emailer) SendPasswordRecovery(email string, name, secret []byte) error {
	from := EmailSender + "@" + e.Domain
	msg := e.BuildPasswordRecovery(email, name, secret)
	return e.SendEmail(e.Server, from, email, e.Auth, msg)
}

func (e *Emailer) BuildPasswordRecovery(email string, name, secret []byte) []byte {
	template := e.EmailVerificationTemplate
	template = bytes.Replace(template, EmailSecretToken, secret, -1)
	template = bytes.Replace(template, EmailNameToken, name, -1)
	template = bytes.Replace(template, EmailDomainToken, []byte(e.Domain), -1)

	from := EmailSender + "@" + e.Domain
	msg := bytes.NewBufferString("To: ")
	msg.Write(name)
	msg.WriteString(" <" + email + ">\r\n")
	msg.WriteString("From: ")
	msg.WriteString(EmailSenderName)
	msg.WriteString(" <" + from + ">\r\n")
	msg.WriteString("Subject: Password Reset\r\n")
	msg.WriteString("\r\n")
	msg.Write(template)
	msg.WriteString("\r\n")

	return msg.Bytes()
}

func (e *Emailer) SendEmail(server, from, to string, auth smtp.Auth, msg []byte) error {
	return smtp.SendMail(e.Server, e.Auth, from, []string{to}, msg)
}
