package helper

import (
	"fmt"
	"net/http"
)

// AppError adalah error yang membawa status HTTP dan pesan untuk klien.
//
// Service cukup RETURN error ini; satu ErrorHandler terpusat (config/app.go)
// yang menerjemahkannya menjadi response. Dengan begitu tidak ada service
// yang menyusun response error sendiri-sendiri.
type AppError struct {
	Status  int
	Message string
	Errors  map[string][]string // rincian per field, khusus validasi (422)
	cause   error               // penyebab asli; hanya untuk log, tidak dikirim ke klien
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

// Unwrap membuat errors.Is / errors.As dapat menembus ke penyebab aslinya.
func (e *AppError) Unwrap() error { return e.cause }

func newAppError(status int, message string) *AppError {
	return &AppError{Status: status, Message: message}
}

func BadRequest(message string) *AppError   { return newAppError(http.StatusBadRequest, message) }
func Unauthorized(message string) *AppError { return newAppError(http.StatusUnauthorized, message) }
func Forbidden(message string) *AppError    { return newAppError(http.StatusForbidden, message) }
func NotFound(message string) *AppError     { return newAppError(http.StatusNotFound, message) }
func Conflict(message string) *AppError     { return newAppError(http.StatusConflict, message) }

// Unprocessable dipakai untuk aturan bisnis yang dilanggar (kuota penuh,
// SKS melebihi batas), yaitu permintaan yang bentuknya benar tapi tidak bisa diproses.
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

// Validation membawa rincian error per field (status 422).
func Validation(errs map[string][]string) *AppError {
	e := newAppError(http.StatusUnprocessableEntity, "Validasi gagal")
	e.Errors = errs
	return e
}

// Internal membungkus error tak terduga (database mati, bug, dst).
// Klien hanya melihat pesan umum; penyebab aslinya masuk ke log.
func Internal(cause error) *AppError {
	e := newAppError(http.StatusInternalServerError, "Terjadi kesalahan pada server")
	e.cause = cause
	return e
}
