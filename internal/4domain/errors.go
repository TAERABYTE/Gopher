package domain

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidCreds   = errors.New("invalid credentials")
	ErrUserExists     = errors.New("user already exists")
	ErrInternalServer = errors.New("internal server error")
	ErrUnauthorized   = errors.New("unauthorized request")
)
