package user

import (
	"uuid"

	"github.com/puriice/waypoint/domain/errors"
)

type User struct {
	ID       uuid.UUID
	Username Username
	Email    Email
	Password Password
}

func NewUser(
	userID string,
	username string,
	email string,
	password string,
) (User, error) {
	userUUID, err := uuid.Parse(userID)

	if err != nil {
		return User{}, errors.NewErrInvalidUUID(userID, "user")
	}

	usrname, err := NewUsername(username)

	if err != nil {
		return User{}, err
	}

	emil, err := NewEmail(email)

	if err != nil {
		return User{}, err
	}

	return User{
		ID:       userUUID,
		Username: usrname,
		Email:    emil,
		Password: NewPassword(password),
	}, nil
}
