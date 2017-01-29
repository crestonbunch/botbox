package api

import (
	"bytes"
	"net/smtp"
)

// EmailSender is the name that email is sent as.
const EmailSender = "mailbot"

// EmailSenderName is appears in the "From" field in an email client.
const EmailSenderName = "Botbox"

// EmailSecretToken is a string inside HTML templates that gets replaced with
// a secret token passed into the emailer interface.
var EmailSecretToken = []byte("{{secret}}")

// EmailNameToken is a string that gets replaced with a user's name in an
// HTML email template.
var EmailNameToken = []byte("{{name}}")

// EmailDomainToken is a string that gets replaced with the site domain in
// an HTML template (e.g. 'example.com').
var EmailDomainToken = []byte("{{domain}}")

// EmailerModel is a generic interface for an emailer that sends emails to
// clients. Can be mocked for testing.
type EmailerModel interface {
	// Send an email with a secret key to verify that the email exists and is
	// valid.
	SendEmailVerification(email string, name, secret []byte) error
	// Send a password recovery email to reset an account's password.
	SendPasswordRecovery(email string, name, secret []byte) error
}

// Emailer is a global implementation of EmailerModel which sends emails via
// SMTP.
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

// SendEmailVerification sends a verification email to the given user, filling
// in {{name}} and {{secret}} in the HTML template from the Emailer struct.
func (e *Emailer) SendEmailVerification(email string, name, secret []byte) error {
	from := EmailSender + "@" + e.Domain
	msg := e.BuildEmailVerification(email, name, secret)
	return e.SendEmail(e.Server, from, email, e.Auth, msg)
}

// BuildEmailVerification builds an email verification body to be send via
// SMTP to a user. It replaces {{name}} and {{secret}} in the HTML body with
// the user name and email verification secret.
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
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.Write(template)
	msg.WriteString("\r\n")

	return msg.Bytes()
}

// SendPasswordRecovery sends a password recovery email to the given email
// address, replacing {{name}} and {{secret}} in the HTML template from the
// Emailer struct.
func (e *Emailer) SendPasswordRecovery(email string, name, secret []byte) error {
	from := EmailSender + "@" + e.Domain
	msg := e.BuildPasswordRecovery(email, name, secret)
	return e.SendEmail(e.Server, from, email, e.Auth, msg)
}

// BuildPasswordRecovery builds an email for recovering a password from the
// HTML template in the Emailer struct by replacing {{name}} and {{secret}} in
// the template with the user's name and password recovery secret.
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
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.Write(template)
	msg.WriteString("\r\n")

	return msg.Bytes()
}

// SendEmail is a helper function for sending email with the given authentication
// method in the Emailer struct.
func (e *Emailer) SendEmail(server, from, to string, auth smtp.Auth, msg []byte) error {
	return smtp.SendMail(e.Server, e.Auth, from, []string{to}, msg)
}
