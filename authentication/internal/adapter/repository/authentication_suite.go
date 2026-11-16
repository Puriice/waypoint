// Package repository holds behaviour shared by every repository adapter,
// such as the contract test suites each implementation must pass.
package repository

import (
	"context"
	"strings"
	"sync"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
)

// AuthenticationFactory returns a ready-to-use repository for a single test.
// Implementations may share storage between calls: the suite generates unique
// usernames, emails and fingerprints so tests never collide.
type AuthenticationFactory func(t *testing.T) repository.Authentication

// RunAuthenticationSuite checks that an implementation of
// repository.Authentication honours the contract defined by the migrations:
// unique emails and usernames, non-empty emails, password hashes longer than
// 32 bytes, one session per (user, device fingerprint), and cascading deletes.
func RunAuthenticationSuite(t *testing.T, newRepo AuthenticationFactory) {
	t.Run("Register", func(t *testing.T) {
		t.Run("creates user, session and refresh claims", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()

			u, session, claims, err := in.register(t, repo)
			require.NoError(t, err)

			assert.NotEqual(t, uuid.Nil(), u.ID)
			assert.Equal(t, in.username, u.Username)
			assert.Equal(t, in.email, u.Email)
			assert.Equal(t, in.password.Expose(), u.Password.Expose())

			assert.NotEqual(t, uuid.Nil(), session.ID)
			assert.Equal(t, in.fingerprint, session.DeviceFingerprint)
			assert.Nil(t, session.MFAVerifiedAt)

			assert.Equal(t, session.ID, claims.SessionID)
			assert.NotEqual(t, uuid.Nil(), claims.FamilyID)
			assert.Equal(t, in.refreshKeyHash.Expose(), claims.Key.Expose())
		})

		t.Run("rejects duplicated email", func(t *testing.T) {
			repo := newRepo(t)
			first := newRegistration()
			_, _, _, err := first.register(t, repo)
			require.NoError(t, err)

			second := newRegistration()
			second.email = first.email

			_, _, _, err = second.register(t, repo)
			require.ErrorIs(t, err, _domain.ErrDuplicatedEmail)
		})

		t.Run("rejects duplicated username", func(t *testing.T) {
			repo := newRepo(t)
			first := newRegistration()
			_, _, _, err := first.register(t, repo)
			require.NoError(t, err)

			second := newRegistration()
			second.username = first.username

			_, _, _, err = second.register(t, repo)
			require.ErrorIs(t, err, _domain.ErrDuplicatedUsername)

			// The whole registration is atomic: no user is left behind.
			_, err = repo.GetByEmail(context.Background(), second.email)
			require.ErrorIs(t, err, _domain.ErrNotFound)
		})

		t.Run("lets exactly one of concurrent registrations with the same email succeed", func(t *testing.T) {
			repo := newRepo(t)
			email := newRegistration().email

			const attempts = 8
			errs := make(chan error, attempts)
			var wg sync.WaitGroup

			for range attempts {
				in := newRegistration()
				in.email = email

				wg.Go(func() {
					_, _, _, err := in.register(t, repo)
					errs <- err
				})
			}

			wg.Wait()
			close(errs)

			succeeded := 0
			for err := range errs {
				if err == nil {
					succeeded++
					continue
				}
				assert.ErrorIs(t, err, _domain.ErrDuplicatedEmail)
			}
			assert.Equal(t, 1, succeeded)
		})

		t.Run("rejects blank email", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			in.email = user.Email("   ")

			_, _, _, err := in.register(t, repo)
			require.ErrorIs(t, err, _domain.ErrInvalidEmail)
		})

		t.Run("rejects password hash of 32 bytes or fewer", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			in.password = user.NewPassword(strings.Repeat("a", 32))

			_, _, _, err := in.register(t, repo)
			require.ErrorIs(t, err, _domain.ErrInvalidPassword)
		})

		t.Run("ignores surrounding whitespace when measuring password hash", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			in.password = user.NewPassword("  " + strings.Repeat("a", 32) + "  ")

			_, _, _, err := in.register(t, repo)
			require.ErrorIs(t, err, _domain.ErrInvalidPassword)
		})
	})

	t.Run("GetByEmail", func(t *testing.T) {
		t.Run("returns the registered user", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			registered, _, _, err := in.register(t, repo)
			require.NoError(t, err)

			got, err := repo.GetByEmail(context.Background(), in.email)
			require.NoError(t, err)

			assert.Equal(t, registered.ID, got.ID)
			assert.Equal(t, 0, got.Version)
			assert.Equal(t, in.username, got.Username)
			assert.Equal(t, in.email, got.Email)
			assert.Equal(t, in.password.Expose(), got.Password.Expose())
			assert.Nil(t, got.EmailVerifiedAt)
			assert.Nil(t, got.DeactivatedAt)
			assert.Nil(t, got.LockedUntil)
			assert.Nil(t, got.QuarantineAt)
			assert.False(t, got.CreatedAt.IsZero())
			assert.False(t, got.UpdatedAt.IsZero())
		})

		t.Run("returns ErrNotFound for unknown email", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetByEmail(context.Background(), newRegistration().email)
			require.ErrorIs(t, err, _domain.ErrNotFound)
		})
	})

	t.Run("NewSession", func(t *testing.T) {
		t.Run("creates a session for a new device", func(t *testing.T) {
			repo := newRepo(t)
			u, first, _, err := newRegistration().register(t, repo)
			require.NoError(t, err)

			fingerprint := uuid.NewV4()
			refreshKeyHash := newHash()

			session, claims, err := repo.NewSession(context.Background(), u, fingerprint, refreshKeyHash)
			require.NoError(t, err)

			assert.NotEqual(t, first.ID, session.ID)
			assert.Equal(t, fingerprint, session.DeviceFingerprint)
			assert.Equal(t, session.ID, claims.SessionID)
			assert.NotEqual(t, uuid.Nil(), claims.FamilyID)
			assert.Equal(t, refreshKeyHash.Expose(), claims.Key.Expose())
		})

		t.Run("rejects a fingerprint the user already has a session for", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			u, _, _, err := in.register(t, repo)
			require.NoError(t, err)

			_, _, err = repo.NewSession(context.Background(), u, in.fingerprint, newHash())
			require.ErrorIs(t, err, _domain.ErrDuplicatedFingerprint)
		})

		t.Run("allows the same fingerprint for different users", func(t *testing.T) {
			repo := newRepo(t)
			first := newRegistration()
			_, _, _, err := first.register(t, repo)
			require.NoError(t, err)

			second := newRegistration()
			second.fingerprint = first.fingerprint

			_, _, _, err = second.register(t, repo)
			require.NoError(t, err)
		})

		t.Run("fails for a user that does not exist", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			ghost, err := user.NewUser(uuid.NewV7().String(), in.username.String(), in.email.String(), in.password.Expose())
			require.NoError(t, err)

			_, _, err = repo.NewSession(context.Background(), ghost, in.fingerprint, newHash())
			require.Error(t, err)
		})
	})

	t.Run("DeleteSession", func(t *testing.T) {
		t.Run("deletes an existing session", func(t *testing.T) {
			repo := newRepo(t)
			_, session, _, err := newRegistration().register(t, repo)
			require.NoError(t, err)

			require.NoError(t, repo.DeleteSession(context.Background(), session.ID))

			err = repo.DeleteSession(context.Background(), session.ID)
			require.ErrorIs(t, err, _domain.ErrNotFound)
		})

		t.Run("frees the fingerprint for a new session", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			u, session, _, err := in.register(t, repo)
			require.NoError(t, err)

			require.NoError(t, repo.DeleteSession(context.Background(), session.ID))

			_, _, err = repo.NewSession(context.Background(), u, in.fingerprint, newHash())
			require.NoError(t, err)
		})

		t.Run("returns ErrNotFound for unknown session", func(t *testing.T) {
			repo := newRepo(t)

			err := repo.DeleteSession(context.Background(), uuid.NewV7())
			require.ErrorIs(t, err, _domain.ErrNotFound)
		})
	})

	t.Run("DeleteUser", func(t *testing.T) {
		t.Run("deletes the user and their sessions", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			u, session, _, err := in.register(t, repo)
			require.NoError(t, err)

			require.NoError(t, repo.DeleteUser(context.Background(), u.ID))

			_, err = repo.GetByEmail(context.Background(), in.email)
			require.ErrorIs(t, err, _domain.ErrNotFound)

			err = repo.DeleteSession(context.Background(), session.ID)
			require.ErrorIs(t, err, _domain.ErrNotFound)
		})

		t.Run("frees the email and username", func(t *testing.T) {
			repo := newRepo(t)
			in := newRegistration()
			u, _, _, err := in.register(t, repo)
			require.NoError(t, err)

			require.NoError(t, repo.DeleteUser(context.Background(), u.ID))

			_, _, _, err = in.register(t, repo)
			require.NoError(t, err)
		})

		t.Run("returns ErrNotFound for unknown user", func(t *testing.T) {
			repo := newRepo(t)

			err := repo.DeleteUser(context.Background(), uuid.NewV7())
			require.ErrorIs(t, err, _domain.ErrNotFound)
		})
	})
}

type registration struct {
	username       user.Username
	email          user.Email
	password       user.Password
	fingerprint    uuid.UUID
	refreshKeyHash user.Password
}

// newRegistration returns valid, unique registration input.
func newRegistration() registration {
	id := strings.ReplaceAll(uuid.NewV4().String(), "-", "")

	return registration{
		username:       user.Username("user_" + id),
		email:          user.Email("user." + id + "@example.com"),
		password:       newHash(),
		fingerprint:    uuid.NewV4(),
		refreshKeyHash: newHash(),
	}
}

func (r registration) register(t *testing.T, repo repository.Authentication) (user.User, domain.Session, refresh.Claims, error) {
	t.Helper()
	return repo.Register(context.Background(), r.username, r.email, r.password, r.fingerprint, r.refreshKeyHash)
}

// newHash returns a unique value long enough to satisfy
// user_password_hash_not_empty (more than 32 bytes).
func newHash() user.Password {
	return user.NewPassword("$argon2id$" + strings.Repeat(uuid.NewV4().String(), 2))
}
