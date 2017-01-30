package api

// ErrMissingName is returned when an input for name is empty.
var ErrMissingName = &HttpError{"Missing name.", 400}

// ErrMissingEmail is returned when an input for email is empty.
var ErrMissingEmail = &HttpError{"Missing email.", 400}

// ErrMissingPassword is returned when an input for password is empty.
var ErrMissingPassword = &HttpError{"Missing password.", 400}

// ErrNameTooLong is returned when an input for name exceeds a valid length.
var ErrNameTooLong = &HttpError{"You name is too long.", 400}

// ErrPasswordTooShort is returned when an input for a password is too short.
var ErrPasswordTooShort = &HttpError{"You password is too short.", 400}

// ErrInvalidPassword is returned when a password was entered incorrectly.
var ErrInvalidPassword = &HttpError{"The password you entered is incorrect.", 400}

// ErrUnknown is returned when an unknown server error is encountered. It should
// be logged for investigated, but not shown to the user.
var ErrUnknown = &HttpError{"An unknown error occurred.", 500}

// ErrEmailInUse is returned when an email is already being used.
var ErrEmailInUse = &HttpError{"That email is already in use!", 400}

// ErrEmailNotFound is returned when an email is not found in the database.
var ErrEmailNotFound = &HttpError{"That email does not exist.", 404}

// ErrUserNotFound is returned when a user is not found in the database.
var ErrUserNotFound = &HttpError{"Specified user was not found.", 404}

// ErrInvalidJson is returned when malformed JSON is received.
var ErrInvalidJson = &HttpError{"Invalid JSON.", 400}

// ErrBotDetected is returned when a user fails the captcha test.
var ErrBotDetected = &HttpError{"You are a bot.", 400}

// ErrInvalidSecret is returned when a bad secret is received.
var ErrInvalidSecret = &HttpError{"Invalid secret.", 400}

// ErrLoginIncorrect is returned when a user provides a bad email/password combo.
var ErrLoginIncorrect = &HttpError{"Email or password was incorrect.", 400}

// ErrMissingParameter is returned when a parameter is missing.
var ErrMissingParameter = &HttpError{"Missing parameter.", 400}

// ErrNotAnInteger is returned when an integer is not received.
var ErrNotAnInteger = &HttpError{"Please provide an integer.", 400}

// HttpError is returned to a user if something goes wrong.
type HttpError struct {
	Message string
	Status  int
}

// Error returns the error message of the HttpError.
func (e *HttpError) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code of the error.
func (e *HttpError) StatusCode() int {
	return e.Status
}
