package repository

import (
	"context"
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
}
