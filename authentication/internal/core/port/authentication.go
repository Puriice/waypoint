package port

import (
	"context"

	"github.com/puriice/waypoint/authentication/internal/core/domain/cmd"
	"github.com/puriice/waypoint/domain/user"
)

type AuthenticationService interface {
	Login(ctx context.Context, command cmd.Login) (user.User, error)
}
