package repository

import (
	"context"
	"uuid"

	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
)

type Authentication interface {
	Login(ctx context.Context, username user.Username, email user.Email, password user.Password, fingerprint uuid.UUID, refreshKeyHash string) (user.User, domain.Session, refresh.Claims, error)
	Register(ctx context.Context, username user.Username, email user.Email, password user.Password, fingerprint uuid.UUID, refreshKeyHash string) (user.User, domain.Session, refresh.Claims, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
