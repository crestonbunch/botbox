package api

import (
	"bytes"
	"testing"
)

func TestEmailerBuildVerificationEmail(t *testing.T) {

	emailer := &Emailer{
		Auth:   nil,
		Server: "smtp.example.com",
		Domain: "example.com",
		EmailVerificationTemplate: []byte("Hi {{secret}} {{name}} {{domain}}"),
	}

	msg := emailer.BuildEmailVerification(
		"user@example.com", []byte("John Doe"), []byte("abcdef123456"),
	)

	exp := bytes.NewBufferString("To: ")
	exp.WriteString("John Doe")
	exp.WriteString(" <user@example.com>\r\n")
	exp.WriteString("From: ")
	exp.WriteString("Botbox")
	exp.WriteString(" <mailbot@example.com>\r\n")
	exp.WriteString("Subject: Email Verification\r\n")
	exp.WriteString("\r\n")
	exp.WriteString("Hi abcdef123456 John Doe example.com")
	exp.WriteString("\r\n")

	if string(msg) != string(exp.Bytes()) {
		t.Log(string(msg))
		t.Error("Verification email was not generated correctl!")
	}
}

func TestEmailerBuildPasswordRecovery(t *testing.T) {

	emailer := &Emailer{
		Auth:   nil,
		Server: "smtp.example.com",
		Domain: "example.com",
		EmailVerificationTemplate: []byte("Hi {{secret}} {{name}} {{domain}}"),
	}

	msg := emailer.BuildPasswordRecovery(
		"user@example.com", []byte("John Doe"), []byte("abcdef123456"),
	)

	exp := bytes.NewBufferString("To: ")
	exp.WriteString("John Doe")
	exp.WriteString(" <user@example.com>\r\n")
	exp.WriteString("From: ")
	exp.WriteString("Botbox")
	exp.WriteString(" <mailbot@example.com>\r\n")
	exp.WriteString("Subject: Password Reset\r\n")
	exp.WriteString("\r\n")
	exp.WriteString("Hi abcdef123456 John Doe example.com")
	exp.WriteString("\r\n")

	if string(msg) != string(exp.Bytes()) {
		t.Log(string(msg))
		t.Error("Verification email was not generated correctl!")
	}
}
