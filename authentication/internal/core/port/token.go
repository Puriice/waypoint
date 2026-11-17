package port

import (
	"context"

	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token"
	"github.com/puriice/waypoint/domain/token/access"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
)

type AuthenticationTokenizer interface {
	Sign(context context.Context, user user.User, session domain.Session, refresh refresh.Claims) (token.Pair, error)
	VerifyAccessToken(context context.Context, token string) (access.Token, error)
	VerifyRefreshToken(context context.Context, token string) (refresh.Token, error)
}
