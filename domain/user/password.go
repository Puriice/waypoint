package user

import (
	"database/sql/driver"
	"fmt"

	"github.com/puriice/waypoint/domain"
)

type Password struct {
	domain.SensitiveString
}

func NewPassword(password string) Password {
	return Password{
		domain.NewSensitiveString(password),
	}
}

func (p *Password) Apply(mutate func(password string) (string, error)) error {
	newPassword, err := mutate(p.Expose())

	if err != nil {
		return err
	}

	p.SensitiveString = domain.NewSensitiveString(newPassword)

	return nil
}

func (p *Password) Scan(src any) error {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case string:
		p.SensitiveString = domain.NewSensitiveString(v)
		return nil
	default:
		return fmt.Errorf("unsupported type %T for Password", src)
	}
}

func (p Password) Value() (driver.Value, error) {
	return p.Expose(), nil
}
