package domain

import "errors"

// Authentication Repository
var (
	ErrDuplicatedEmail       = errors.New("The email already exists.")
	ErrDuplicatedUsername    = errors.New("The username already exists.")
	ErrDuplicatedFingerprint = errors.New("The fingerprint already exists for the user.")
	ErrInvalidEmail          = errors.New("The email is invalid.")
	ErrInvalidPassword       = errors.New("The hashed password is invalid.")
)

// General Repository
var (
	ErrUnknown  = errors.New("Unknown error had occured!")
	ErrNotFound = errors.New("Resource not found.")
)

// Tokenizer
var (
	ErrInvalidSigningKey  = errors.New("The token signing key is invalid.")
	ErrInvalidTokenClaims = errors.New("The token claims cannot be encoded.")
	ErrFailToSignToken    = errors.New("Failed to sign the token.")
	ErrUnknownToken       = errors.New("Unknown error type")
	ErrInvalidToken       = errors.New("The token is invalid.")
	ErrTokenExpired       = errors.New("The token has expired.")
)
