package user

import (
	"time"
	"uuid"

	"github.com/puriice/waypoint/domain/errors"
)

type User struct {
	ID              uuid.UUID
	Version         int
	Username        Username
	Email           Email
	Password        Password
	EmailVerifiedAt *time.Time
	DeactivatedAt   *time.Time
	LockedUntil     *time.Time
	QuarantineAt    *time.Time
	UpdatedAt       time.Time
	CreatedAt       time.Time
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
