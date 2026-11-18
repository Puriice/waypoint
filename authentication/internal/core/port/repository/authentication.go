package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
)

type Authentication interface {
	Register(ctx context.Context, username user.Username, email user.Email, password user.Password, fingerprint uuid.UUID, refreshKeyHash user.Password) (user.User, domain.Session, refresh.Claims, error)
	NewSession(ctx context.Context, user user.User, fingerprint uuid.UUID, refreshKeyHash user.Password) (domain.Session, refresh.Claims, error)
	GetByEmail(ctx context.Context, email user.Email) (user.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	Lock(ctx context.Context, user user.User, until time.Time) (user.User, error)
	Block(ctx context.Context, user user.User, unlockTokenHash user.Password) (user.User, error)
	Unblock(ctx context.Context, user user.User) (user.User, error)
}

type LoginCounter interface {
	Count(ctx context.Context, user user.User) (int, error)
}
