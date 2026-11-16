package cmd

import (
	"uuid"

	"github.com/puriice/waypoint/domain/errors"
	"github.com/puriice/waypoint/domain/user"
)

type Register struct {
	username    string
	email       string
	password    string
	fingerprint string
}

func NewRegister(username, email, password, fingerprint string) Register {
	return Register{
		username:    username,
		email:       email,
		password:    password,
		fingerprint: fingerprint,
	}
}

func (c Register) Username() (user.Username, error) {
	return user.NewUsername(c.username)
}

func (c Register) Email() (user.Email, error) {
	return user.NewEmail(c.email)
}

func (c Register) Password() user.Password {
	return user.NewPassword(c.password)
}

func (c Register) Fingerprint() (uuid.UUID, error) {
	fingerprint, err := uuid.Parse(c.fingerprint)

	if err != nil {
		return uuid.Nil(), errors.NewErrInvalidUUID(c.fingerprint, "device fingerprint register command")
	}

	return fingerprint, nil
}
