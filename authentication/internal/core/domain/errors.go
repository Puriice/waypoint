package domain

import "errors"

var (
	ErrDuplicatedEmail       = errors.New("The email already exists.")
	ErrDuplicatedUsername    = errors.New("The username already exists.")
	ErrDuplicatedFingerprint = errors.New("The fingerprint already exists for the user.")
	ErrInvalidEmail          = errors.New("The email is invalid.")
	ErrInvalidPassword       = errors.New("The hashed password is invalid.")
	ErrUnknown               = errors.New("Unknown error had occured!")
)

var (
	ErrNotFound = errors.New("Resource not found.")
)

var (
	ErrInvalidSigningKey  = errors.New("The token signing key is invalid.")
	ErrInvalidTokenClaims = errors.New("The token claims cannot be encoded.")
	ErrFailToSignToken    = errors.New("Failed to sign the token.")
)
