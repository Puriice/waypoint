package port

import (
	"context"

	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/user"
)

type AuthenticationTokenizer interface {
	Sign(context context.Context, user user.User, persona string) (domain.AccessToken, domain.RefreshToken, error)
	Verify(context context.Context, access domain.AccessToken, refresh domain.RefreshToken) error
	RotateToken(context context.Context, refresh domain.RefreshToken, persona string) (domain.AccessToken, domain.RefreshToken, error)
}
