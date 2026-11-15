package user

import "github.com/puriice/waypoint/domain"

type Password struct {
	domain.SensitiveString
}

func NewPassword(password string) Password {
	return Password{
		domain.NewSensitiveString(password),
	}
}
