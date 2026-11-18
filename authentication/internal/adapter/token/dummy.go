package token

import (
	"context"
	"fmt"
	"sync"
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

// DummyTokenizer is a port.AuthenticationTokenizer for tests. It does no
// cryptography: Sign hands out opaque strings ("dummy-access-1",
// "dummy-refresh-1", ...) and remembers the claims behind each one, and the
// Verify methods look them up again. Expiry, issuer, clock and errors are all
// configurable, and every Sign call is recorded for assertions.
//
// It is safe for concurrent use.
type DummyTokenizer struct {
	mu sync.Mutex

	issuer          string
	accessTTL       time.Duration
	refreshTTL      time.Duration
	accessExpiresAt *time.Time
	refreshExpiry   *time.Time
	now             func() time.Time

	signErr          error
	verifyAccessErr  error
	verifyRefreshErr error

	n       int
	access  map[string]access.Token
	refresh map[string]refresh.Token
	signed  []DummySignCall
}

// DummySignCall records the arguments of one successful Sign call and the
// pair it returned.
type DummySignCall struct {
	User    user.User
	Session domain.Session
	Claims  refresh.Claims
	Pair    token.Pair
}

type DummyOption func(*DummyTokenizer)

// WithDummyIssuer sets the iss claim. Default: "dummy".
func WithDummyIssuer(issuer string) DummyOption {
	return func(t *DummyTokenizer) { t.issuer = issuer }
}

// WithDummyAccessTTL sets how long access tokens live. Default: 5 minutes.
// A negative TTL issues tokens that are already expired.
func WithDummyAccessTTL(ttl time.Duration) DummyOption {
	return func(t *DummyTokenizer) { t.accessTTL = ttl }
}

// WithDummyRefreshTTL sets how long refresh tokens live. Default: 30 days.
func WithDummyRefreshTTL(ttl time.Duration) DummyOption {
	return func(t *DummyTokenizer) { t.refreshTTL = ttl }
}

// WithDummyAccessExpiresAt gives every access token this exact expiry,
// overriding the access TTL.
func WithDummyAccessExpiresAt(at time.Time) DummyOption {
	return func(t *DummyTokenizer) { t.accessExpiresAt = &at }
}

// WithDummyRefreshExpiresAt gives every refresh token this exact expiry,
// overriding the refresh TTL.
func WithDummyRefreshExpiresAt(at time.Time) DummyOption {
	return func(t *DummyTokenizer) { t.refreshExpiry = &at }
}

// WithDummyClock replaces time.Now for issuing and checking expiry.
func WithDummyClock(now func() time.Time) DummyOption {
	return func(t *DummyTokenizer) { t.now = now }
}

// WithDummySignError makes Sign fail with err.
func WithDummySignError(err error) DummyOption {
	return func(t *DummyTokenizer) { t.signErr = err }
}

// WithDummyVerifyAccessError makes VerifyAccessToken fail with err.
func WithDummyVerifyAccessError(err error) DummyOption {
	return func(t *DummyTokenizer) { t.verifyAccessErr = err }
}

// WithDummyVerifyRefreshError makes VerifyRefreshToken fail with err.
func WithDummyVerifyRefreshError(err error) DummyOption {
	return func(t *DummyTokenizer) { t.verifyRefreshErr = err }
}

var _ port.AuthenticationTokenizer = (*DummyTokenizer)(nil)

func NewDummyTokenizer(opts ...DummyOption) *DummyTokenizer {
	t := &DummyTokenizer{
		issuer:     "dummy",
		accessTTL:  5 * time.Minute,
		refreshTTL: 30 * 24 * time.Hour,
		now:        time.Now,
		access:     make(map[string]access.Token),
		refresh:    make(map[string]refresh.Token),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *DummyTokenizer) Sign(_ context.Context, u user.User, session domain.Session, claims refresh.Claims) (token.Pair, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.signErr != nil {
		return token.Pair{}, t.signErr
	}

	now := t.now()
	t.n++

	accessSigned := fmt.Sprintf("dummy-access-%d", t.n)
	refreshSigned := fmt.Sprintf("dummy-refresh-%d", t.n)

	accessToken := access.NewToken(
		t.registeredClaims(now, t.accessTTL, t.accessExpiresAt),
		u,
		session,
		domain.NewSensitiveString(accessSigned),
	)
	refreshToken := refresh.NewToken(
		t.registeredClaims(now, t.refreshTTL, t.refreshExpiry),
		claims,
		domain.NewSensitiveString(refreshSigned),
	)

	t.access[accessSigned] = accessToken
	t.refresh[refreshSigned] = refreshToken

	pair := token.NewPair(accessToken, refreshToken)
	t.signed = append(t.signed, DummySignCall{User: u, Session: session, Claims: claims, Pair: pair})

	return pair, nil
}

func (t *DummyTokenizer) VerifyAccessToken(_ context.Context, signed string) (access.Token, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.verifyAccessErr != nil {
		return access.Token{}, t.verifyAccessErr
	}

	tok, ok := t.access[signed]
	if !ok {
		return access.Token{}, fmt.Errorf("%w: unknown dummy access token", _domain.ErrInvalidToken)
	}

	if err := t.checkExpiry(tok.ExpiresAt); err != nil {
		return access.Token{}, err
	}

	return tok, nil
}

func (t *DummyTokenizer) VerifyRefreshToken(_ context.Context, signed string) (refresh.Token, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.verifyRefreshErr != nil {
		return refresh.Token{}, t.verifyRefreshErr
	}

	tok, ok := t.refresh[signed]
	if !ok {
		return refresh.Token{}, fmt.Errorf("%w: unknown dummy refresh token", _domain.ErrInvalidToken)
	}

	if err := t.checkExpiry(tok.ExpiresAt); err != nil {
		return refresh.Token{}, err
	}

	return tok, nil
}

// SetSignError changes the Sign error mid-test; nil restores success.
func (t *DummyTokenizer) SetSignError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.signErr = err
}

// SetClock changes the clock mid-test, e.g. to move past an expiry.
func (t *DummyTokenizer) SetClock(now func() time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.now = now
}

// Revoke makes a previously issued access or refresh token fail verification
// with ErrInvalidToken.
func (t *DummyTokenizer) Revoke(signed string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.access, signed)
	delete(t.refresh, signed)
}

// Signed returns a copy of every successful Sign call, oldest first.
func (t *DummyTokenizer) Signed() []DummySignCall {
	t.mu.Lock()
	defer t.mu.Unlock()

	return append([]DummySignCall(nil), t.signed...)
}

// LastSigned returns the most recent successful Sign call. ok is false when
// Sign has not succeeded yet.
func (t *DummyTokenizer) LastSigned() (call DummySignCall, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.signed) == 0 {
		return DummySignCall{}, false
	}

	return t.signed[len(t.signed)-1], true
}

func (t *DummyTokenizer) registeredClaims(now time.Time, ttl time.Duration, fixed *time.Time) jwt.RegisteredClaims {
	expiresAt := now.Add(ttl)
	if fixed != nil {
		expiresAt = *fixed
	}

	return jwt.RegisteredClaims{
		Issuer:    t.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
}

// checkExpiry must be called with t.mu held.
func (t *DummyTokenizer) checkExpiry(expiresAt *jwt.NumericDate) error {
	if expiresAt != nil && !t.now().Before(expiresAt.Time) {
		return fmt.Errorf("%w: dummy token expired at %s", _domain.ErrTokenExpired, expiresAt.Time)
	}

	return nil
}
