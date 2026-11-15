package repository

import (
	"context"

	"github.com/puriice/waypoint/domain/user"
)

type Authentication interface {
	Login(ctx context.Context, email, password string) (user.User, error)
	Register(ctx context.Context, email, password string) (user.User, error)
}
