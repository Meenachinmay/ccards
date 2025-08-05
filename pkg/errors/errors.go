package errors

import "errors"

var (
	ErrCompanyExists       = errors.New("company already exists")
	ErrUserExists          = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrCompanySuspended    = errors.New("company account suspended")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrBadRequest          = errors.New("bad request")
	ErrInternalServerError = errors.New("internal server error")
)
