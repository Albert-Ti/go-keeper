package service

import "errors"

var (
	ErrAlreadyExists     = errors.New("user already exist")
	ErrInvalidPassword   = errors.New("incorrect password")
	ErrInvalidCodeEmail  = errors.New("invalid confirmation code")
	ErrEmailNotConfirmed = errors.New("email has not been confirmed")
	ErrNoRows            = errors.New("no rows")
	ErrNotFound          = errors.New("NotFound")
	ErrPasswordReused    = errors.New("password was used before")
)
