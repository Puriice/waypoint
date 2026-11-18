// Package inmemory provides repository adapters backed by process memory.
// They mirror the constraints declared in migrations/ so they can stand in
// for the PostgreSQL adapters in tests and local development.
package inmemory

import (
	"context"
	"strings"
	"sync"
	"time"
	"uuid"

	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
)

// minPasswordHashBytes mirrors user_password_hash_not_empty:
// octet_length(btrim(password_hash)) > 32.
const minPasswordHashBytes = 32

type refreshToken struct {
	familyID       uuid.UUID
	currentKeyHash user.Password
	expiredAt      time.Time
	rotatedAt      time.Time
}

type session struct {
	userID uuid.UUID
	domain.Session
	refreshToken refreshToken
}

type AuthenticationRepository struct {
	mu                 sync.RWMutex
	users              map[uuid.UUID]user.User
	sessions           map[uuid.UUID]session
	refreshKeyDuration time.Duration
	now                func() time.Time
}

func NewAuthentication(refreshKeyDuration time.Duration) repository.Authentication {
	return &AuthenticationRepository{
		users:              make(map[uuid.UUID]user.User),
		sessions:           make(map[uuid.UUID]session),
		refreshKeyDuration: refreshKeyDuration,
		now:                time.Now,
	}
}

func (r *AuthenticationRepository) GetByEmail(_ context.Context, email user.Email) (user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}

	return user.User{}, _domain.ErrNotFound
}

func (r *AuthenticationRepository) Register(_ context.Context, username user.Username, email user.Email, password user.Password, fingerprint uuid.UUID, refreshKeyHash user.Password) (user.User, domain.Session, refresh.Claims, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.validateNewUser(username, email, password); err != nil {
		return user.User{}, domain.Session{}, refresh.Claims{}, err
	}

	now := r.now()
	u := user.User{
		ID:        uuid.NewV7(),
		Username:  username,
		Email:     email,
		Password:  password,
		UpdatedAt: now,
		CreatedAt: now,
	}

	// Validation is complete and newSession cannot fail for a brand-new user,
	// so storing the user first keeps Register atomic.
	r.users[u.ID] = u

	s, claims, err := r.newSession(now, u.ID, fingerprint, refreshKeyHash)
	if err != nil {
		delete(r.users, u.ID)
		return user.User{}, domain.Session{}, refresh.Claims{}, err
	}

	return u, s, claims, nil
}

func (r *AuthenticationRepository) NewSession(_ context.Context, u user.User, fingerprint uuid.UUID, refreshKeyHash user.Password) (domain.Session, refresh.Claims, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[u.ID]; !ok {
		// Mirrors the sessions.user_id foreign key.
		return domain.Session{}, refresh.Claims{}, _domain.ErrNotFound
	}

	return r.newSession(r.now(), u.ID, fingerprint, refreshKeyHash)
}

func (r *AuthenticationRepository) DeleteUser(_ context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[userID]; !ok {
		return _domain.ErrNotFound
	}

	delete(r.users, userID)

	// Mirrors ON DELETE CASCADE on sessions (and through them, refresh_tokens).
	for id, s := range r.sessions {
		if s.userID == userID {
			delete(r.sessions, id)
		}
	}

	return nil
}

func (r *AuthenticationRepository) DeleteSession(_ context.Context, sessionID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[sessionID]; !ok {
		return _domain.ErrNotFound
	}

	delete(r.sessions, sessionID)

	return nil
}

// validateNewUser mirrors the CHECK and UNIQUE constraints on users.
// The caller must hold r.mu.
func (r *AuthenticationRepository) validateNewUser(username user.Username, email user.Email, password user.Password) error {
	if len(btrim(email.String())) == 0 {
		return _domain.ErrInvalidEmail
	}

	if len(btrim(password.Expose())) <= minPasswordHashBytes {
		return _domain.ErrInvalidPassword
	}

	for _, u := range r.users {
		if u.Email == email {
			return _domain.ErrDuplicatedEmail
		}
		if u.Username == username {
			return _domain.ErrDuplicatedUsername
		}
	}

	return nil
}

// newSession stores a session and its refresh token. The caller must hold
// r.mu and have checked that userID exists.
func (r *AuthenticationRepository) newSession(now time.Time, userID, fingerprint uuid.UUID, refreshKeyHash user.Password) (domain.Session, refresh.Claims, error) {
	// Mirrors device_fingerprint_unique: UNIQUE(user_id, device_fingerprint).
	for _, s := range r.sessions {
		if s.userID == userID && s.DeviceFingerprint == fingerprint {
			return domain.Session{}, refresh.Claims{}, _domain.ErrDuplicatedFingerprint
		}
	}

	s := session{
		userID: userID,
		Session: domain.Session{
			ID:                uuid.NewV7(),
			DeviceFingerprint: fingerprint,
			UpdatedAt:         now,
			CreatedAt:         now,
		},
		refreshToken: refreshToken{
			familyID:       uuid.NewV7(),
			currentKeyHash: refreshKeyHash,
			expiredAt:      now.Add(r.refreshKeyDuration),
			rotatedAt:      now,
		},
	}

	r.sessions[s.ID] = s

	claims := refresh.Claims{
		SessionID: s.ID,
		FamilyID:  s.refreshToken.familyID,
		Key:       refreshKeyHash,
	}

	return s.Session, claims, nil
}

// btrim matches PostgreSQL's btrim(text), which strips spaces only.
func btrim(s string) string {
	return strings.Trim(s, " ")
}

func (r *AuthenticationRepository) Lock(ctx context.Context, user user.User, until time.Time) (user.User, error) {
	panic("unimplement")
}

func (r *AuthenticationRepository) Block(ctx context.Context, user user.User, unlockKeyHash user.Password) (user.User, error) {
	panic("unimplement")
}

func (r *AuthenticationRepository) Unblock(ctx context.Context, user user.User) (user.User, error) {
	panic("unimplement")
}
