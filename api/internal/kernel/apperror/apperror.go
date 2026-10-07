package apperror

import (
	"errors"
	"maps"
	"strconv"
	"strings"
)

const secondsPerMinute = 60

type FieldError struct {
	Field   string
	Code    string
	Message string
}

type Error struct {
	code       Code
	params     map[string]string
	fields     []FieldError
	retryAfter int
	cause      error
}

func New(code Code) *Error {
	return &Error{code: code}
}

func Validation(fields ...FieldError) *Error {
	return &Error{code: CodeValidationFailed, fields: fields}
}

var ErrOutsideScope = errors.New("data belongs to another organization")

func NotFound() *Error {
	return New(CodeNotFound)
}

func OutsideScope() *Error {
	return NotFound().WithCause(ErrOutsideScope)
}

func Internal(cause error) *Error {
	return &Error{code: CodeInternal, cause: cause}
}

func RateLimited(retryAfterSeconds int) *Error {
	return retryLater(CodeRateLimited, retryAfterSeconds)
}

func AccountLocked(retryAfterSeconds int) *Error {
	return retryLater(CodeAccountLocked, retryAfterSeconds)
}

func retryLater(code Code, retryAfterSeconds int) *Error {
	minutes := (retryAfterSeconds + secondsPerMinute - 1) / secondsPerMinute
	limited := &Error{code: code, retryAfter: retryAfterSeconds}
	return limited.WithParam("n", strconv.Itoa(minutes))
}

func Unavailable(retryAfterSeconds int) *Error {
	return &Error{code: CodeServiceUnavailable, retryAfter: retryAfterSeconds}
}

func From(err error) *Error {
	if appErr, ok := errors.AsType[*Error](err); ok {
		return appErr
	}
	return Internal(err)
}

func (e *Error) WithParam(name, value string) *Error {
	derived := *e
	derived.params = maps.Clone(e.params)
	if derived.params == nil {
		derived.params = map[string]string{}
	}
	derived.params[name] = value
	return &derived
}

func (e *Error) WithCause(cause error) *Error {
	derived := *e
	derived.cause = cause
	return &derived
}

func (e *Error) Error() string {
	if e.cause == nil {
		return string(e.code)
	}
	return string(e.code) + ": " + e.cause.Error()
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Code() Code {
	return e.code
}

func (e *Error) Kind() Kind {
	return lookup(e.code).kind
}

func (e *Error) Status() int {
	return lookup(e.code).status
}

func (e *Error) Fields() []FieldError {
	return e.fields
}

func (e *Error) RetryAfter() int {
	return e.retryAfter
}

func (e *Error) Message() string {
	message := lookup(e.code).message
	for name, value := range e.params {
		message = strings.ReplaceAll(message, "{"+name+"}", value)
	}
	return message
}
