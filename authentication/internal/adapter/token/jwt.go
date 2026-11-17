package token

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"uuid"

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

type key struct {
	AccessToken  *rsa.PrivateKey
	RefreshToken *rsa.PrivateKey
}

type JWTTokenizer struct {
	issuer string
	ttl    ttl
	key    key
}

func NewJWTTokenizer(issuer string, accessTTL, refreshTTL time.Duration, accessKey, refreshKey *rsa.PrivateKey) port.AuthenticationTokenizer {
	return JWTTokenizer{
		issuer: issuer,
		ttl: ttl{
			AccessToken:  accessTTL,
			RefreshToken: refreshTTL,
		},
		key: key{
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
		"",
	)

	err := accessToken.Sign(t.key.AccessToken)

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
		"",
	)

	refreshToken.ExpiresAt = jwt.NewNumericDate(now.Add(t.ttl.RefreshToken))
	err = refreshToken.Sign(t.key.RefreshToken)

	if err != nil {
		return token.Pair{}, mapSignError(err)
	}

	return token.NewPair(accessToken, refreshToken), nil
}

func (t JWTTokenizer) VerifyAccessToken(context context.Context, token string) (access.Token, error) {
	var claims access.Token

	if err := t.parse(token, &claims, &t.key.AccessToken.PublicKey); err != nil {
		return access.Token{}, err
	}

	userID, err := uuid.Parse(claims.Subject)

	if err != nil {
		return access.Token{}, fmt.Errorf("%w: invalid subject: %w", _domain.ErrInvalidToken, err)
	}

	sessionID, err := uuid.Parse(claims.SessionID)

	if err != nil {
		return access.Token{}, fmt.Errorf("%w: invalid session id: %w", _domain.ErrInvalidToken, err)
	}

	return access.NewToken(
		claims.RegisteredClaims,
		user.User{ID: userID, Email: user.Email(claims.Email)},
		domain.Session{ID: sessionID},
		domain.NewSensitiveString(token),
	), nil
}

func (t JWTTokenizer) VerifyRefreshToken(context context.Context, token string) (refresh.Token, error) {
	var claims refresh.Token

	if err := t.parse(token, &claims, &t.key.RefreshToken.PublicKey); err != nil {
		return refresh.Token{}, err
	}

	if claims.SessionID == uuid.Nil() || claims.FamilyID == uuid.Nil() || claims.Key.Expose() == "" {
		return refresh.Token{}, fmt.Errorf("%w: missing session id, family id or key", _domain.ErrInvalidToken)
	}

	return refresh.NewToken(claims.RegisteredClaims, claims.Claims, domain.NewSensitiveString(token)), nil
}

// parse verifies the RS256 signature with key and validates the registered
// claims (issuer, expiry, issued-at) before decoding into claims.
func (t JWTTokenizer) parse(token string, claims jwt.Claims, key *rsa.PublicKey) error {
	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(*jwt.Token) (any, error) { return key, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(t.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)

	if err != nil {
		return mapVerifyError(err)
	}

	return nil
}

func mapVerifyError(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return fmt.Errorf("%w: %w", _domain.ErrTokenExpired, err)
	case errors.Is(err, jwt.ErrTokenMalformed),
		errors.Is(err, jwt.ErrTokenUnverifiable),
		errors.Is(err, jwt.ErrTokenSignatureInvalid),
		errors.Is(err, jwt.ErrTokenRequiredClaimMissing),
		errors.Is(err, jwt.ErrTokenInvalidIssuer),
		errors.Is(err, jwt.ErrTokenUsedBeforeIssued),
		errors.Is(err, jwt.ErrTokenNotValidYet),
		errors.Is(err, jwt.ErrTokenInvalidClaims):
		return fmt.Errorf("%w: %w", _domain.ErrInvalidToken, err)
	}

	return fmt.Errorf("%w: %w", _domain.ErrUnknownToken, err)
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
