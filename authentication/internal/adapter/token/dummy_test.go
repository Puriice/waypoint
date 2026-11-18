package token_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_token "github.com/puriice/waypoint/authentication/internal/adapter/token"
	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/port"
)

func newDummy(_ *testing.T, cfg tokenizerConfig) port.AuthenticationTokenizer {
	return _token.NewDummyTokenizer(
		_token.WithDummyIssuer(cfg.issuer),
		_token.WithDummyAccessTTL(cfg.accessTTL),
		_token.WithDummyRefreshTTL(cfg.refreshTTL),
	)
}

// Dummy-only test controls; the shared contract lives in suite_test.go.

func TestDummyTokenizerControls(t *testing.T) {
	ctx := context.Background()

	t.Run("issues distinct tokens per call and records them", func(t *testing.T) {
		tok := _token.NewDummyTokenizer()

		first, second := signDefault(t, tok), signDefault(t, tok)

		assert.NotEqual(t, first.access, second.access)
		assert.Len(t, tok.Signed(), 2)

		last, ok := tok.LastSigned()
		require.True(t, ok)
		assert.Equal(t, second.access, last.Pair.AccessToken.SignedToken().Expose())
	})

	t.Run("uses the clock for TTLs", func(t *testing.T) {
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		tok := _token.NewDummyTokenizer(
			_token.WithDummyClock(func() time.Time { return now }),
			_token.WithDummyAccessTTL(time.Minute),
			_token.WithDummyRefreshTTL(time.Hour),
		)
		u, session, claims := tokenInputs(t)

		pair, err := tok.Sign(ctx, u, session, claims)
		require.NoError(t, err)

		assert.Equal(t, now, pair.AccessToken.IssuedAt.Time.UTC())
		assert.Equal(t, now.Add(time.Minute), pair.AccessToken.ExpiresAt.Time.UTC())
		assert.Equal(t, now.Add(time.Hour), pair.RefreshToken.ExpiresAt.Time.UTC())
	})

	t.Run("uses fixed expiry dates when given", func(t *testing.T) {
		accessAt := time.Date(2030, 6, 1, 12, 0, 0, 0, time.UTC)
		refreshAt := time.Date(2031, 6, 1, 12, 0, 0, 0, time.UTC)
		tok := _token.NewDummyTokenizer(
			_token.WithDummyAccessExpiresAt(accessAt),
			_token.WithDummyRefreshExpiresAt(refreshAt),
		)
		u, session, claims := tokenInputs(t)

		pair, err := tok.Sign(ctx, u, session, claims)
		require.NoError(t, err)

		assert.Equal(t, accessAt, pair.AccessToken.ExpiresAt.Time.UTC())
		assert.Equal(t, refreshAt, pair.RefreshToken.ExpiresAt.Time.UTC())
	})

	t.Run("expires tokens when the clock moves past expiry", func(t *testing.T) {
		now := time.Now()
		tok := _token.NewDummyTokenizer(
			_token.WithDummyClock(func() time.Time { return now }),
			_token.WithDummyAccessTTL(time.Minute),
		)
		pair := signDefault(t, tok)

		tok.SetClock(func() time.Time { return now.Add(2 * time.Minute) })

		_, err := tok.VerifyAccessToken(ctx, pair.access)
		require.ErrorIs(t, err, _domain.ErrTokenExpired)
	})

	t.Run("rejects tokens from another instance and revoked tokens", func(t *testing.T) {
		tok := _token.NewDummyTokenizer()
		pair := signDefault(t, tok)

		_, err := _token.NewDummyTokenizer().VerifyAccessToken(ctx, pair.access)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)

		tok.Revoke(pair.access)
		_, err = tok.VerifyAccessToken(ctx, pair.access)
		require.ErrorIs(t, err, _domain.ErrInvalidToken)
	})

	t.Run("returns configured errors", func(t *testing.T) {
		boom := errors.New("boom")
		tok := _token.NewDummyTokenizer(
			_token.WithDummySignError(_domain.ErrInvalidSigningKey),
			_token.WithDummyVerifyAccessError(boom),
			_token.WithDummyVerifyRefreshError(boom),
		)
		u, session, claims := tokenInputs(t)

		_, err := tok.Sign(ctx, u, session, claims)
		require.ErrorIs(t, err, _domain.ErrInvalidSigningKey)
		_, ok := tok.LastSigned()
		assert.False(t, ok, "failed Sign calls are not recorded")

		tok.SetSignError(nil)
		_, err = tok.Sign(ctx, u, session, claims)
		require.NoError(t, err)

		_, err = tok.VerifyAccessToken(ctx, "anything")
		require.ErrorIs(t, err, boom)
		_, err = tok.VerifyRefreshToken(ctx, "anything")
		require.ErrorIs(t, err, boom)
	})
}
