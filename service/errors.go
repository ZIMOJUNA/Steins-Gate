package service

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrEmailExists         = errors.New("email already exists")
	ErrAccountNotFound     = errors.New("account not found")
	ErrAccountDisabled     = errors.New("account disabled")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrInvalidCode         = errors.New("invalid verification code")
	ErrCodeExpired         = errors.New("verification code expired")
	ErrCodeTooManyAttempts = errors.New("too many verification attempts")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrSaveNotFound        = errors.New("player data not found")
)
