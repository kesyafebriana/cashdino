package model

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
)

// appError wraps a sentinel error while preserving a custom message.
// This allows errors.Is to match the sentinel, and .Error() to return the original message.
type appError struct {
	sentinel error
	msg      string
}

func (e *appError) Error() string { return e.msg }
func (e *appError) Unwrap() error { return e.sentinel }

func ValidationErr(msg string) error {
	return &appError{sentinel: ErrValidation, msg: msg}
}

func NotFoundErr(msg string) error {
	return &appError{sentinel: ErrNotFound, msg: msg}
}
