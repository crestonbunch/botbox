package api

var ErrMissingName = &HttpError{"Missing name.", 400}
var ErrMissingEmail = &HttpError{"Missing email.", 400}
var ErrMissingPassword = &HttpError{"Missing password.", 400}
var ErrNameTooLong = &HttpError{"You name is too long.", 400}
var ErrPasswordTooShort = &HttpError{"You password is too short.", 400}
var ErrUnknown = &HttpError{"An unknown error occurred.", 500}
var ErrEmailInUse = &HttpError{"That email is already in use!", 400}
var ErrEmailNotFound = &HttpError{"That email does not exist.", 400}
var ErrInvalidJson = &HttpError{"Invalid JSON.", 400}
var ErrBotDetected = &HttpError{"You are a bot.", 400}
var ErrInvalidSecret = &HttpError{"Invalid secret.", 400}
var ErrUserNotFound = &HttpError{"Specified user was not found.", 400}

// An HTTP error that can be returned to a user if something goes wrong.
type HttpError struct {
	Message string
	Status  int
}

func (e *HttpError) Error() string {
	return e.Message
}

func (e *HttpError) StatusCode() int {
	return e.Status
}
