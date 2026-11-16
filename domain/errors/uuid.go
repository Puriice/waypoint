package errors

import "fmt"

type ErrInvalidUUID struct {
	uuid string
	role string
}

func NewErrInvalidUUID(uuid, role string) ErrInvalidUUID {
	return ErrInvalidUUID{
		uuid: uuid,
		role: role,
	}
}

func (e ErrInvalidUUID) Error() string {
	return fmt.Sprintf("\"%s\" is not a valid %s's uuid", e.uuid, e.role)
}

func (e ErrInvalidUUID) Is(other error) bool {
	switch other.(type) {
	case ErrInvalidUUID:
		return true
	default:
		return false
	}
}
