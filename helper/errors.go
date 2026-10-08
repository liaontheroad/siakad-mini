package helper

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Status  int
	Message string
	Errors  map[string][]string 
	cause   error               
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.cause }

func newAppError(status int, message string) *AppError {
	return &AppError{Status: status, Message: message}
}

func BadRequest(message string) *AppError   { return newAppError(http.StatusBadRequest, message) }
func Unauthorized(message string) *AppError { return newAppError(http.StatusUnauthorized, message) }
func Forbidden(message string) *AppError    { return newAppError(http.StatusForbidden, message) }
func NotFound(message string) *AppError     { return newAppError(http.StatusNotFound, message) }
func Conflict(message string) *AppError     { return newAppError(http.StatusConflict, message) }

func Unprocessable(message string) *AppError {
	return newAppError(http.StatusUnprocessableEntity, message)
}

func TooManyRequests(message string) *AppError {
	return newAppError(http.StatusTooManyRequests, message)
}

func UnsupportedMediaType(message string) *AppError {
	return newAppError(http.StatusUnsupportedMediaType, message)
}

func ServiceUnavailable(message string) *AppError {
	return newAppError(http.StatusServiceUnavailable, message)
}

func Validation(errs map[string][]string) *AppError {
	e := newAppError(http.StatusUnprocessableEntity, "Validasi gagal")
	e.Errors = errs
	return e
}

func Internal(cause error) *AppError {
	e := newAppError(http.StatusInternalServerError, "Terjadi kesalahan pada server")
	e.cause = cause
	return e
}
