package token

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/port"
	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token"
	"github.com/puriice/waypoint/domain/token/access"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
)

type ttl struct {
	AccessToken  time.Duration
	RefreshToken time.Duration
}

type privateKey struct {
	AccessToken  *rsa.PrivateKey
	RefreshToken *rsa.PrivateKey
}

type JWTTokenizer struct {
	issuer     string
	ttl        ttl
	privateKey privateKey
}

func NewJWTTokenizer(issuer string, accessTTL, refreshTTL time.Duration, accessKey, refreshKey *rsa.PrivateKey) port.AuthenticationTokenizer {
	return JWTTokenizer{
		issuer: issuer,
		ttl: ttl{
			AccessToken:  accessTTL,
			RefreshToken: refreshTTL,
		},
		privateKey: privateKey{
			AccessToken:  accessKey,
			RefreshToken: refreshKey,
		},
	}
}

func (t JWTTokenizer) Sign(context context.Context, user user.User, session domain.Session, claims refresh.Claims) (token.Pair, error) {
	now := time.Now()
	jwtNow := jwt.NewNumericDate(now)

	accessToken := access.NewToken(
		jwt.RegisteredClaims{
			Issuer:    t.issuer,
			IssuedAt:  jwtNow,
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl.AccessToken)),
		},
		user,
		session,
	)

	err := accessToken.Sign(t.privateKey.AccessToken)

	if err != nil {
		return token.Pair{}, mapSignError(err)
	}

	refreshToken := refresh.NewToken(
		jwt.RegisteredClaims{
			Issuer:    t.issuer,
			IssuedAt:  jwtNow,
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl.RefreshToken)),
		},
		claims,
	)

	refreshToken.ExpiresAt = jwt.NewNumericDate(now.Add(t.ttl.RefreshToken))
	err = refreshToken.Sign(t.privateKey.RefreshToken)

	if err != nil {
		return token.Pair{}, mapSignError(err)
	}

	return token.NewPair(accessToken, refreshToken), nil
}

func mapSignError(err error) error {
	var marshalerErr *json.MarshalerError
	var unsupportedTypeErr *json.UnsupportedTypeError
	var unsupportedValueErr *json.UnsupportedValueError

	switch {
	case errors.Is(err, jwt.ErrInvalidKeyType),
		errors.Is(err, jwt.ErrInvalidKey),
		errors.Is(err, jwt.ErrTokenUnverifiable):
		return fmt.Errorf("%w: %w", _domain.ErrInvalidSigningKey, err)
	case errors.As(err, &marshalerErr),
		errors.As(err, &unsupportedTypeErr),
		errors.As(err, &unsupportedValueErr):
		return fmt.Errorf("%w: %w", _domain.ErrInvalidTokenClaims, err)
	}

	return fmt.Errorf("%w: %w", _domain.ErrFailToSignToken, err)
}

func (t JWTTokenizer) Verify(context context.Context, token access.Token) error {
	panic("unimplemented")
}

func (t JWTTokenizer) RotateToken(context context.Context, refresh refresh.Token) (token.Pair, error) {
	panic("unimplemented")
}
