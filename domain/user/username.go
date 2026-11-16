package user

import (
	"database/sql/driver"
	"fmt"
)

type Username string

func NewUsername(username string) (Username, error) {
	return Username(username), nil
}

func (u Username) String() string {
	return string(u)
}

func (u *Username) Scan(src any) error {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case string:
		username, err := NewUsername(v)
		if err != nil {
			return err
		}
		*u = username
		return nil
	default:
		return fmt.Errorf("unsupported type %T for Username", src)
	}
}

func (u Username) Value() (driver.Value, error) {
	return string(u), nil
}
