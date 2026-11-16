package cmd

import (
	"uuid"

	"github.com/puriice/waypoint/domain/errors"
	"github.com/puriice/waypoint/domain/user"
)

type Login struct {
	email       string
	password    string
	fingerprint string
}

func NewLogin(email, password, fingerprint string) Login {
	return Login{
		email:       email,
		password:    password,
		fingerprint: fingerprint,
	}
}

func (c Login) Email() (user.Email, error) {
	return user.NewEmail(c.email)
}

func (c Login) Password() user.Password {
	return user.NewPassword(c.password)
}

func (c Login) Fingerprint() (uuid.UUID, error) {
	fingerprint, err := uuid.Parse(c.fingerprint)

	if err != nil {
		return uuid.Nil(), errors.NewErrInvalidUUID(c.fingerprint, "device fingerprint register command")
	}

	return fingerprint, nil
}
