package token_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_token "github.com/puriice/waypoint/authentication/internal/adapter/token"
	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
)

const issuer = "waypoint-test"

var (
	accessKey  = mustKey()
	refreshKey = mustKey()
	otherKey   = mustKey()
)

func mustKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return key
}

type fixture struct {
	user    user.User
	session domain.Session
	claims  refresh.Claims
	pair    token.Pair
}

func sign(t *testing.T, issuer string, accessTTL, refreshTTL time.Duration, accessKey, refreshKey *rsa.PrivateKey) fixture {
	t.Helper()

	u, err := user.NewUser(uuid.NewV7().String(), "alice", "alice@example.com", "hash")
	require.NoError(t, err)

	session := domain.Session{ID: uuid.NewV7(), DeviceFingerprint: uuid.NewV4()}
	claims := refresh.Claims{SessionID: session.ID, FamilyID: uuid.NewV7(), Key: user.NewPassword("refresh-key")}

	tokenizer := _token.NewJWTTokenizer(issuer, accessTTL, refreshTTL, accessKey, refreshKey)
	pair, err := tokenizer.Sign(context.Background(), u, session, claims)
	require.NoError(t, err)

	return fixture{user: u, session: session, claims: claims, pair: pair}
}

func TestVerifyAccessToken(t *testing.T) {
	tokenizer := _token.NewJWTTokenizer(issuer, time.Minute, time.Hour, accessKey, refreshKey)

	t.Run("returns the claims of a valid token", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, time.Hour, accessKey, refreshKey)

		got, err := tokenizer.VerifyAccessToken(context.Background(), f.pair.AccessToken.SignedToken().Expose())
		require.NoError(t, err)

		assert.Equal(t, f.user.ID.String(), got.Subject)
		assert.Equal(t, f.user.Email.String(), got.Email)
		assert.Equal(t, f.session.ID.String(), got.SessionID)
		assert.Equal(t, issuer, got.Issuer)
		assert.Equal(t, f.pair.AccessToken.SignedToken().Expose(), got.SignedToken().Expose())
	})

	t.Run("rejects a subject that is not a uuid", func(t *testing.T) {
		signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss":        issuer,
			"sub":        "alice",
			"session_id": uuid.NewV7().String(),
			"exp":        time.Now().Add(time.Minute).Unix(),
		}).SignedString(accessKey)
		require.NoError(t, err)

		_, err = tokenizer.VerifyAccessToken(context.Background(), signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects an expired token", func(t *testing.T) {
		f := sign(t, issuer, -time.Minute, time.Hour, accessKey, refreshKey)

		_, err := tokenizer.VerifyAccessToken(context.Background(), f.pair.AccessToken.SignedToken().Expose())
		require.ErrorIs(t, err, _domain.ErrTokenExpired)
	})

	t.Run("rejects a token from another issuer", func(t *testing.T) {
		f := sign(t, "someone-else", time.Minute, time.Hour, accessKey, refreshKey)

		_, err := tokenizer.VerifyAccessToken(context.Background(), f.pair.AccessToken.SignedToken().Expose())
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a token signed with another key", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, time.Hour, otherKey, refreshKey)

		_, err := tokenizer.VerifyAccessToken(context.Background(), f.pair.AccessToken.SignedToken().Expose())
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a refresh token", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, time.Hour, accessKey, refreshKey)

		_, err := tokenizer.VerifyAccessToken(context.Background(), f.pair.RefreshToken.SignedToken().Expose())
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a tampered payload", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, time.Hour, accessKey, refreshKey)
		other := sign(t, issuer, time.Minute, time.Hour, accessKey, refreshKey)

		// Swap in another valid token's payload but keep the original signature.
		parts := strings.Split(f.pair.AccessToken.SignedToken().Expose(), ".")
		parts[1] = strings.Split(other.pair.AccessToken.SignedToken().Expose(), ".")[1]

		_, err := tokenizer.VerifyAccessToken(context.Background(), strings.Join(parts, "."))
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a token using a different algorithm", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss":        issuer,
			"sub":        uuid.NewV7().String(),
			"session_id": uuid.NewV7().String(),
			"iat":        time.Now().Unix(),
			"exp":        time.Now().Add(time.Minute).Unix(),
		}

		hs256, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
		require.NoError(t, err)
		none, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		for name, signed := range map[string]string{"HS256": hs256, "none": none} {
			_, err := tokenizer.VerifyAccessToken(context.Background(), signed)
			assert.ErrorIs(t, err, _domain.ErrInvalidToken, name)
		}
	})

	t.Run("rejects a token without an expiry", func(t *testing.T) {
		signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss":        issuer,
			"sub":        uuid.NewV7().String(),
			"session_id": uuid.NewV7().String(),
		}).SignedString(accessKey)
		require.NoError(t, err)

		_, err = tokenizer.VerifyAccessToken(context.Background(), signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a token without subject or session id", func(t *testing.T) {
		signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss": issuer,
			"exp": time.Now().Add(time.Minute).Unix(),
		}).SignedString(accessKey)
		require.NoError(t, err)

		_, err = tokenizer.VerifyAccessToken(context.Background(), signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects garbage", func(t *testing.T) {
		_, err := tokenizer.VerifyAccessToken(context.Background(), "not.a.jwt")
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})
}

func TestVerifyRefreshToken(t *testing.T) {
	tokenizer := _token.NewJWTTokenizer(issuer, time.Minute, time.Hour, accessKey, refreshKey)

	t.Run("returns the claims of a valid token", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, time.Hour, accessKey, refreshKey)

		got, err := tokenizer.VerifyRefreshToken(context.Background(), f.pair.RefreshToken.SignedToken().Expose())
		require.NoError(t, err)

		assert.Equal(t, f.claims.SessionID, got.SessionID)
		assert.Equal(t, f.claims.FamilyID, got.FamilyID)
		assert.Equal(t, f.claims.Key.Expose(), got.Key.Expose())
		assert.Equal(t, issuer, got.Issuer)
		assert.Equal(t, f.pair.RefreshToken.SignedToken().Expose(), got.SignedToken().Expose())
	})

	t.Run("encodes the key as a plain string", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, time.Hour, accessKey, refreshKey)

		payload, err := base64.RawURLEncoding.DecodeString(strings.Split(f.pair.RefreshToken.SignedToken().Expose(), ".")[1])
		require.NoError(t, err)

		var claims map[string]any
		require.NoError(t, json.Unmarshal(payload, &claims))
		assert.Equal(t, f.claims.Key.Expose(), claims["k"])
	})

	t.Run("rejects an expired token", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, -time.Minute, accessKey, refreshKey)

		_, err := tokenizer.VerifyRefreshToken(context.Background(), f.pair.RefreshToken.SignedToken().Expose())
		require.ErrorIs(t, err, _domain.ErrTokenExpired)
	})

	t.Run("rejects an access token", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, time.Hour, accessKey, refreshKey)

		_, err := tokenizer.VerifyRefreshToken(context.Background(), f.pair.AccessToken.SignedToken().Expose())
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a token signed with another key", func(t *testing.T) {
		f := sign(t, issuer, time.Minute, time.Hour, accessKey, otherKey)

		_, err := tokenizer.VerifyRefreshToken(context.Background(), f.pair.RefreshToken.SignedToken().Expose())
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a token without session, family or key", func(t *testing.T) {
		signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss": issuer,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
		}).SignedString(refreshKey)
		require.NoError(t, err)

		_, err = tokenizer.VerifyRefreshToken(context.Background(), signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a malformed session id", func(t *testing.T) {
		signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss":        issuer,
			"exp":        time.Now().Add(time.Hour).Unix(),
			"session_id": "not-a-uuid",
		}).SignedString(refreshKey)
		require.NoError(t, err)

		_, err = tokenizer.VerifyRefreshToken(context.Background(), signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})
}
