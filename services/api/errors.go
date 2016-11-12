package api

import (
	"errors"
)

var UnknownError = errors.New("An unknown error occurred. Please try again.")
var UserNotFound = errors.New("Specified user was not found.")
var BotDetected = errors.New("You are a bot.")
var MissingNewAccountField = errors.New("Missing username, email, or password.")
var UsernameExists = errors.New("That username already exists!")
var EmailInUse = errors.New("That email is already in use!")
var InvalidJson = errors.New("Invalid JSON.")
var UsernameTooLong = errors.New("You username is too long.")
var PasswordTooShort = errors.New("You password is too short.")
