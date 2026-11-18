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
	"github.com/puriice/waypoint/authentication/internal/core/port"
)

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

func newJWT(_ *testing.T, cfg tokenizerConfig) port.AuthenticationTokenizer {
	return _token.NewJWTTokenizer(cfg.issuer, cfg.accessTTL, cfg.refreshTTL, accessKey, refreshKey)
}

// JWT-specific behaviour; the shared contract lives in suite_test.go.

func TestJWTTokenizerVerifyAccessToken(t *testing.T) {
	ctx := context.Background()
	tok := newJWT(t, defaultConfig())

	t.Run("rejects a token signed with another key", func(t *testing.T) {
		other := _token.NewJWTTokenizer(suiteIssuer, time.Minute, time.Hour, otherKey, refreshKey)

		_, err := tok.VerifyAccessToken(ctx, signDefault(t, other).access)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a tampered payload", func(t *testing.T) {
		first, second := signDefault(t, tok), signDefault(t, tok)

		// Swap in another valid token's payload but keep the original signature.
		parts := strings.Split(first.access, ".")
		parts[1] = strings.Split(second.access, ".")[1]

		_, err := tok.VerifyAccessToken(ctx, strings.Join(parts, "."))
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a token using a different algorithm", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss":        suiteIssuer,
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
			_, err := tok.VerifyAccessToken(ctx, signed)
			assert.ErrorIs(t, err, _domain.ErrInvalidToken, name)
		}
	})

	t.Run("rejects a token without an expiry", func(t *testing.T) {
		signed := signClaims(t, accessKey, jwt.MapClaims{
			"iss":        suiteIssuer,
			"sub":        uuid.NewV7().String(),
			"session_id": uuid.NewV7().String(),
		})

		_, err := tok.VerifyAccessToken(ctx, signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a token without subject or session id", func(t *testing.T) {
		signed := signClaims(t, accessKey, jwt.MapClaims{
			"iss": suiteIssuer,
			"exp": time.Now().Add(time.Minute).Unix(),
		})

		_, err := tok.VerifyAccessToken(ctx, signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a subject that is not a uuid", func(t *testing.T) {
		signed := signClaims(t, accessKey, jwt.MapClaims{
			"iss":        suiteIssuer,
			"sub":        "alice",
			"session_id": uuid.NewV7().String(),
			"exp":        time.Now().Add(time.Minute).Unix(),
		})

		_, err := tok.VerifyAccessToken(ctx, signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})
}

func TestJWTTokenizerVerifyRefreshToken(t *testing.T) {
	ctx := context.Background()
	tok := newJWT(t, defaultConfig())

	t.Run("rejects a token signed with another key", func(t *testing.T) {
		other := _token.NewJWTTokenizer(suiteIssuer, time.Minute, time.Hour, accessKey, otherKey)

		_, err := tok.VerifyRefreshToken(ctx, signDefault(t, other).refresh)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("encodes the key as a plain string", func(t *testing.T) {
		payload, err := base64.RawURLEncoding.DecodeString(strings.Split(signDefault(t, tok).refresh, ".")[1])
		require.NoError(t, err)

		var claims map[string]any
		require.NoError(t, json.Unmarshal(payload, &claims))
		assert.Equal(t, "refresh-key", claims["k"])
	})

	t.Run("rejects a token without session, family or key", func(t *testing.T) {
		signed := signClaims(t, refreshKey, jwt.MapClaims{
			"iss": suiteIssuer,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
		})

		_, err := tok.VerifyRefreshToken(ctx, signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("rejects a malformed session id", func(t *testing.T) {
		signed := signClaims(t, refreshKey, jwt.MapClaims{
			"iss":        suiteIssuer,
			"exp":        time.Now().Add(time.Hour).Unix(),
			"session_id": "not-a-uuid",
		})

		_, err := tok.VerifyRefreshToken(ctx, signed)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})
}

func signClaims(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()

	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	require.NoError(t, err)

	return signed
}
