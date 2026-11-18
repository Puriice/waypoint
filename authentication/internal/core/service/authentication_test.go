package service_test

import (
	"context"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/puriice/waypoint/authentication/internal/adapter/repository/inmemory"
	_token "github.com/puriice/waypoint/authentication/internal/adapter/token"
	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/domain/cmd"
	"github.com/puriice/waypoint/authentication/internal/core/port"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
	"github.com/puriice/waypoint/authentication/internal/core/service"
	"github.com/puriice/waypoint/domain/app"
	"github.com/puriice/waypoint/domain/serror"
	_user "github.com/puriice/waypoint/domain/user"
)

const (
	testEmail    = "alice@example.com"
	testPassword = "correct horse battery staple"
)

var invalidCredential = serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)

func TestLogin(t *testing.T) {
	t.Run("returns the user and signed tokens for valid credentials", func(t *testing.T) {
		d := newDeps(t)
		registered := d.registerUser(t)
		fingerprint := uuid.NewV4()

		u, pair, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, fingerprint.String()))
		require.NoError(t, err)

		assert.Equal(t, registered.ID, u.ID)
		signed, ok := d.tokenizer.LastSigned()
		require.True(t, ok)
		assert.Equal(t, signed.Pair, pair)
		assert.Equal(t, registered.ID, signed.User.ID)
		assert.Equal(t, fingerprint, signed.Session.DeviceFingerprint)
		assert.Equal(t, signed.Session.ID, signed.Claims.SessionID)
		assert.Equal(t, d.key.lastKey(), signed.Claims.Key.Expose(), "token must carry the plain refresh key, not its hash")
	})

	t.Run("rejects a wrong password and applies penalties", func(t *testing.T) {
		d := newDeps(t)
		d.registerUser(t)

		_, _, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, "wrong password", uuid.NewV4().String()))
		require.ErrorIs(t, err, invalidCredential)

		assert.Eventually(t, func() bool { return d.counter.calls() == 1 }, time.Second, 10*time.Millisecond,
			"a wrong password must count towards the penalty")
		assert.Empty(t, d.tokenizer.Signed())
	})

	t.Run("rejects an unknown email after a dummy verification", func(t *testing.T) {
		d := newDeps(t)

		_, _, err := d.service().Login(newContext(), cmd.NewLogin("nobody@example.com", testPassword, uuid.NewV4().String()))
		require.ErrorIs(t, err, invalidCredential)

		assert.Equal(t, 1, d.hasher.verifies(), "unknown email must still cost one password verification")
		assert.Zero(t, d.counter.calls())
	})

	t.Run("rejects a repository failure after a dummy verification", func(t *testing.T) {
		d := newDeps(t)
		d.repo = failingRepository{}

		_, _, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, uuid.NewV4().String()))
		require.ErrorIs(t, err, invalidCredential)

		assert.Equal(t, 1, d.hasher.verifies())
	})

	t.Run("rejects an invalid fingerprint", func(t *testing.T) {
		d := newDeps(t)
		d.registerUser(t)

		_, _, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, "not-a-uuid"))
		require.ErrorIs(t, err, invalidCredential)
	})

	t.Run("rejects a malformed email", func(t *testing.T) {
		d := newDeps(t)

		_, _, err := d.service().Login(newContext(), cmd.NewLogin("not-an-email", testPassword, uuid.NewV4().String()))
		require.ErrorIs(t, err, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidEmail))
	})

	t.Run("rejects a locked account even with the right password", func(t *testing.T) {
		d := newDeps(t)
		d.registerUser(t)
		until := time.Now().Add(time.Hour)
		d.repo = alteringRepository{Authentication: d.repo, alter: func(u *_user.User) { u.LockedUntil = &until }}

		_, _, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, uuid.NewV4().String()))
		require.ErrorIs(t, err, invalidCredential)
		assert.Empty(t, d.tokenizer.Signed())
	})

	t.Run("rejects a quarantined account even with the right password", func(t *testing.T) {
		d := newDeps(t)
		d.registerUser(t)
		at := time.Now()
		d.repo = alteringRepository{Authentication: d.repo, alter: func(u *_user.User) { u.QuarantineAt = &at }}

		_, _, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, uuid.NewV4().String()))
		require.ErrorIs(t, err, invalidCredential)
		assert.Empty(t, d.tokenizer.Signed())
	})

	t.Run("rejects a device that already has a session", func(t *testing.T) {
		d := newDeps(t)
		d.registerUser(t)
		fingerprint := uuid.NewV4().String()

		_, _, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, fingerprint))
		require.NoError(t, err)

		_, _, err = d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, fingerprint))
		require.ErrorIs(t, err, invalidCredential)
	})

	t.Run("deletes the session when signing fails", func(t *testing.T) {
		d := newDeps(t)
		d.registerUser(t)
		d.tokenizer.SetSignError(_domain.ErrFailToSignToken)
		fingerprint := uuid.NewV4().String()

		_, _, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, fingerprint))
		require.ErrorIs(t, err, invalidCredential)

		// Once the session is gone, the same device can log in again.
		d.tokenizer.SetSignError(nil)
		assert.Eventually(t, func() bool {
			_, _, err := d.service().Login(newContext(), cmd.NewLogin(testEmail, testPassword, fingerprint))
			return err == nil
		}, time.Second, 10*time.Millisecond)
	})
}

// deps wires the service with an in-memory repository and recording fakes.
type deps struct {
	repo      repository.Authentication
	counter   *fakeCounter
	tokenizer *_token.DummyTokenizer
	key       *fakeKey
	hasher    *recordingHasher
}

func newDeps(t *testing.T) *deps {
	t.Helper()

	return &deps{
		repo:      inmemory.NewAuthentication(time.Hour),
		counter:   &fakeCounter{count: 1},
		tokenizer: _token.NewDummyTokenizer(),
		key:       &fakeKey{},
		hasher:    &recordingHasher{},
	}
}

func (d *deps) service() port.AuthenticationService {
	return service.NewAuthenticationService(d.repo, d.counter, d.tokenizer, d.key, d.hasher, fakeEmail{})
}

// registerUser stores a user whose password hash matches testPassword.
func (d *deps) registerUser(t *testing.T) _user.User {
	t.Helper()

	hash, err := d.hasher.Hash(testPassword)
	require.NoError(t, err)
	refreshHash, err := d.hasher.Hash("initial-refresh-key-padding-to-pass-length-checks")
	require.NoError(t, err)

	u, _, _, err := d.repo.Register(
		context.Background(),
		_user.Username("alice"),
		_user.Email(testEmail),
		_user.NewPassword(hash),
		uuid.NewV4(),
		_user.NewPassword(refreshHash),
	)
	require.NoError(t, err)

	return u
}

func newContext() app.Context {
	return app.Context{Context: context.Background(), RequestID: uuid.NewV4()}
}

// recordingHasher "hashes" by prefixing, padded so hashes pass the
// repository's minimum length, and counts Verify calls.
type recordingHasher struct {
	mu          sync.Mutex
	verifyCalls int
}

const hashPrefix = "hashed-with-a-prefix-long-enough-for-checks:"

func (h *recordingHasher) Hash(input string) (string, error) { return hashPrefix + input, nil }

func (h *recordingHasher) Verify(input, hash string) bool {
	h.mu.Lock()
	h.verifyCalls++
	h.mu.Unlock()

	return hash == hashPrefix+input
}

func (h *recordingHasher) verifies() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.verifyCalls
}

type fakeKey struct {
	mu   sync.Mutex
	n    int
	last string
}

func (k *fakeKey) Generate(int) _user.Password {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.n++
	k.last = "refresh-key-" + string(rune('a'+k.n))

	return _user.NewPassword(k.last)
}

func (k *fakeKey) lastKey() string {
	k.mu.Lock()
	defer k.mu.Unlock()

	return k.last
}

// fakeCounter reports a fixed failed-login count. With the default
// forgiven window (3), a count of 1 never reaches Lock or Block.
type fakeCounter struct {
	mu    sync.Mutex
	count int
	n     int
}

func (c *fakeCounter) Count(context.Context, _user.User) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.n++

	return c.count, nil
}

func (c *fakeCounter) calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.n
}

type fakeEmail struct{}

func (fakeEmail) SendUnlockLink(context.Context, _user.Email, _user.Password) error { return nil }

// failingRepository simulates an unavailable database.
type failingRepository struct {
	repository.Authentication
}

func (failingRepository) GetByEmail(context.Context, _user.Email) (_user.User, error) {
	return _user.User{}, _domain.ErrUnknown
}

// alteringRepository modifies users returned by GetByEmail, e.g. to mark
// them locked, since the in-memory Lock and Block are not implemented yet.
type alteringRepository struct {
	repository.Authentication
	alter func(*_user.User)
}

func (r alteringRepository) GetByEmail(ctx context.Context, email _user.Email) (_user.User, error) {
	u, err := r.Authentication.GetByEmail(ctx, email)
	if err == nil {
		r.alter(&u)
	}

	return u, err
}
