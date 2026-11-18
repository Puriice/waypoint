package token_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/port"
	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
)

const suiteIssuer = "waypoint-test"

// tokenizerConfig is what every implementation must honour.
type tokenizerConfig struct {
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func defaultConfig() tokenizerConfig {
	return tokenizerConfig{issuer: suiteIssuer, accessTTL: time.Minute, refreshTTL: time.Hour}
}

// tokenizerFactory builds an implementation from cfg. Two calls with the same
// cfg need not accept each other's tokens; calls with different issuers must not.
type tokenizerFactory func(t *testing.T, cfg tokenizerConfig) port.AuthenticationTokenizer

func TestJWTTokenizerContract(t *testing.T) {
	runTokenizerSuite(t, newJWT)
}

func TestDummyTokenizerContract(t *testing.T) {
	runTokenizerSuite(t, newDummy)
}

// runTokenizerSuite checks the behaviour every port.AuthenticationTokenizer
// must share, so the dummy can stand in for the JWT tokenizer in tests.
func runTokenizerSuite(t *testing.T, newTokenizer tokenizerFactory) {
	ctx := context.Background()

	t.Run("Sign", func(t *testing.T) {
		t.Run("returns tokens with the issuer and TTL-based timestamps", func(t *testing.T) {
			cfg := defaultConfig()
			tok := newTokenizer(t, cfg)
			u, session, claims := tokenInputs(t)

			before := time.Now()
			pair, err := tok.Sign(ctx, u, session, claims)
			require.NoError(t, err)

			assert.NotEmpty(t, pair.AccessToken.SignedToken().Expose())
			assert.NotEmpty(t, pair.RefreshToken.SignedToken().Expose())
			assert.NotEqual(t, pair.AccessToken.SignedToken().Expose(), pair.RefreshToken.SignedToken().Expose())

			assert.Equal(t, cfg.issuer, pair.AccessToken.Issuer)
			assert.Equal(t, cfg.issuer, pair.RefreshToken.Issuer)

			// JWT timestamps have one-second precision.
			assert.WithinDuration(t, before, pair.AccessToken.IssuedAt.Time, 2*time.Second)
			assert.WithinDuration(t, before.Add(cfg.accessTTL), pair.AccessToken.ExpiresAt.Time, 2*time.Second)
			assert.WithinDuration(t, before.Add(cfg.refreshTTL), pair.RefreshToken.ExpiresAt.Time, 2*time.Second)
		})

		t.Run("puts the user, session and claims into the tokens", func(t *testing.T) {
			tok := newTokenizer(t, defaultConfig())
			u, session, claims := tokenInputs(t)

			pair, err := tok.Sign(ctx, u, session, claims)
			require.NoError(t, err)

			assert.Equal(t, u.ID.String(), pair.AccessToken.Subject)
			assert.Equal(t, u.Email.String(), pair.AccessToken.Email)
			assert.Equal(t, session.ID.String(), pair.AccessToken.SessionID)
			assert.Equal(t, claims.SessionID, pair.RefreshToken.SessionID)
			assert.Equal(t, claims.FamilyID, pair.RefreshToken.FamilyID)
			assert.Equal(t, claims.Key.Expose(), pair.RefreshToken.Key.Expose())
		})
	})

	t.Run("VerifyAccessToken", func(t *testing.T) {
		t.Run("returns the claims of a token it signed", func(t *testing.T) {
			tok := newTokenizer(t, defaultConfig())
			u, session, claims := tokenInputs(t)
			pair, err := tok.Sign(ctx, u, session, claims)
			require.NoError(t, err)

			got, err := tok.VerifyAccessToken(ctx, pair.AccessToken.SignedToken().Expose())
			require.NoError(t, err)

			assert.Equal(t, u.ID.String(), got.Subject)
			assert.Equal(t, u.Email.String(), got.Email)
			assert.Equal(t, session.ID.String(), got.SessionID)
			assert.Equal(t, suiteIssuer, got.Issuer)
			assert.Equal(t, pair.AccessToken.SignedToken().Expose(), got.SignedToken().Expose())
		})

		t.Run("rejects an expired token", func(t *testing.T) {
			cfg := defaultConfig()
			cfg.accessTTL = -time.Minute
			tok := newTokenizer(t, cfg)
			pair := signDefault(t, tok)

			_, err := tok.VerifyAccessToken(ctx, pair.access)
			require.ErrorIs(t, err, _domain.ErrTokenExpired)
		})

		t.Run("rejects a refresh token", func(t *testing.T) {
			tok := newTokenizer(t, defaultConfig())
			pair := signDefault(t, tok)

			_, err := tok.VerifyAccessToken(ctx, pair.refresh)
			require.ErrorIs(t, err, _domain.ErrInvalidToken)
		})

		t.Run("rejects a token from a tokenizer with another issuer", func(t *testing.T) {
			other := defaultConfig()
			other.issuer = "someone-else"
			pair := signDefault(t, newTokenizer(t, other))

			_, err := newTokenizer(t, defaultConfig()).VerifyAccessToken(ctx, pair.access)
			require.ErrorIs(t, err, _domain.ErrInvalidToken)
		})

		t.Run("rejects garbage", func(t *testing.T) {
			tok := newTokenizer(t, defaultConfig())

			for _, input := range []string{"", "not.a.jwt", "garbage"} {
				_, err := tok.VerifyAccessToken(ctx, input)
				assert.ErrorIs(t, err, _domain.ErrInvalidToken, "input %q", input)
			}
		})
	})

	t.Run("VerifyRefreshToken", func(t *testing.T) {
		t.Run("returns the claims of a token it signed", func(t *testing.T) {
			tok := newTokenizer(t, defaultConfig())
			u, session, claims := tokenInputs(t)
			pair, err := tok.Sign(ctx, u, session, claims)
			require.NoError(t, err)

			got, err := tok.VerifyRefreshToken(ctx, pair.RefreshToken.SignedToken().Expose())
			require.NoError(t, err)

			assert.Equal(t, claims.SessionID, got.SessionID)
			assert.Equal(t, claims.FamilyID, got.FamilyID)
			assert.Equal(t, claims.Key.Expose(), got.Key.Expose())
			assert.Equal(t, suiteIssuer, got.Issuer)
			assert.Equal(t, pair.RefreshToken.SignedToken().Expose(), got.SignedToken().Expose())
		})

		t.Run("rejects an expired token", func(t *testing.T) {
			cfg := defaultConfig()
			cfg.refreshTTL = -time.Minute
			tok := newTokenizer(t, cfg)
			pair := signDefault(t, tok)

			_, err := tok.VerifyRefreshToken(ctx, pair.refresh)
			require.ErrorIs(t, err, _domain.ErrTokenExpired)
		})

		t.Run("keeps refresh tokens valid after the access token expires", func(t *testing.T) {
			cfg := defaultConfig()
			cfg.accessTTL = -time.Minute
			tok := newTokenizer(t, cfg)
			pair := signDefault(t, tok)

			_, err := tok.VerifyRefreshToken(ctx, pair.refresh)
			require.NoError(t, err)
		})

		t.Run("rejects an access token", func(t *testing.T) {
			tok := newTokenizer(t, defaultConfig())
			pair := signDefault(t, tok)

			_, err := tok.VerifyRefreshToken(ctx, pair.access)
			require.ErrorIs(t, err, _domain.ErrInvalidToken)
		})

		t.Run("rejects a token from a tokenizer with another issuer", func(t *testing.T) {
			other := defaultConfig()
			other.issuer = "someone-else"
			pair := signDefault(t, newTokenizer(t, other))

			_, err := newTokenizer(t, defaultConfig()).VerifyRefreshToken(ctx, pair.refresh)
			require.ErrorIs(t, err, _domain.ErrInvalidToken)
		})

		t.Run("rejects garbage", func(t *testing.T) {
			tok := newTokenizer(t, defaultConfig())

			for _, input := range []string{"", "not.a.jwt", "garbage"} {
				_, err := tok.VerifyRefreshToken(ctx, input)
				assert.ErrorIs(t, err, _domain.ErrInvalidToken, "input %q", input)
			}
		})
	})
}

type signedPair struct {
	access  string
	refresh string
}

func signDefault(t *testing.T, tok port.AuthenticationTokenizer) signedPair {
	t.Helper()

	u, session, claims := tokenInputs(t)
	pair, err := tok.Sign(context.Background(), u, session, claims)
	require.NoError(t, err)

	return signedPair{
		access:  pair.AccessToken.SignedToken().Expose(),
		refresh: pair.RefreshToken.SignedToken().Expose(),
	}
}

func tokenInputs(t *testing.T) (user.User, domain.Session, refresh.Claims) {
	t.Helper()

	u, err := user.NewUser(uuid.NewV7().String(), "alice", "alice@example.com", "hash")
	require.NoError(t, err)

	session := domain.Session{ID: uuid.NewV7(), DeviceFingerprint: uuid.NewV4()}
	claims := refresh.Claims{SessionID: session.ID, FamilyID: uuid.NewV7(), Key: user.NewPassword("refresh-key")}

	return u, session, claims
}
