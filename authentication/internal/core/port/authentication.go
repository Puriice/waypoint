package port

import (
	"github.com/puriice/waypoint/authentication/internal/core/domain/cmd"
	"github.com/puriice/waypoint/domain/app"
	"github.com/puriice/waypoint/domain/token"
	"github.com/puriice/waypoint/domain/user"
)

type AuthenticationService interface {
	Register(ctx app.Context, command cmd.Register) (user.User, token.Pair, error)
	Login(ctx app.Context, command cmd.Login) (user.User, token.Pair, error)
}
