package errors

type ErrorStatus int64

const (
	ErrorStatusUnspecified ErrorStatus = iota
	ErrorStatusNotFound
	ErrorStatusUnauthenticated
	ErrorStatusInvalidArgument
	ErrorStatusForbidden
)

const (
	msgNotFound        = "not found"
	msgUnauthenticated = "unauthenticated"
	msgInvalidArgument = "invalid argument"
	msgForbidden       = "forbidden"

	msgInternalServerError = "internal server error"
)

type Error struct {
	err     error
	status  ErrorStatus
	message string
}

func (e *Error) Error() string {
	if e.err != nil {
		return e.err.Error()
	}

	return ""
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Status() ErrorStatus {
	return e.status
}

func WrapError(err error, code ErrorStatus, msg string) *Error {
	return &Error{
		err:     err,
		status:  code,
		message: msg,
	}
}

func WrapNotFound(err error) *Error {
	return WrapError(err, ErrorStatusNotFound, msgNotFound)
}

func WrapUnauthenticated(err error) *Error {
	return WrapError(err, ErrorStatusUnauthenticated, msgUnauthenticated)
}

func WrapForbidden(err error) *Error {
	return WrapError(err, ErrorStatusForbidden, msgForbidden)
}

func WrapInvalidArgument(err error) *Error {
	return WrapError(err, ErrorStatusInvalidArgument, msgInvalidArgument)
}
