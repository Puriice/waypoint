package postgres

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token/refresh"
	"github.com/puriice/waypoint/domain/user"
	_user "github.com/puriice/waypoint/domain/user"
)

type AuthenticationRepository struct {
	db                 *pgxpool.Pool
	refreshKeyDuration time.Duration
}

func NewAuthentication(db *pgxpool.Pool, refreshKeyDuration time.Duration) repository.Authentication {
	return AuthenticationRepository{
		db:                 db,
		refreshKeyDuration: refreshKeyDuration,
	}
}

func (r AuthenticationRepository) GetByEmail(ctx context.Context, email _user.Email) (_user.User, error) {
	var user _user.User

	err := r.db.QueryRow(
		ctx,
		`
			SELECT
				id,
				version,
				email,
				username,
				password_hash,
				email_verified_at,
				deactivated_at,
				locked_until,
				quarantine_at,
				updated_at,
				created_at
			FROM users
			WHERE email = $1
			LIMIT 1;
		`,
		email,
	).Scan(
		&user.ID,
		&user.Version,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.DeactivatedAt,
		&user.LockedUntil,
		&user.QuarantineAt,
		&user.UpdatedAt,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return _user.User{}, _domain.ErrNotFound
		}
		return _user.User{}, _domain.ErrUnknown
	}

	return user, nil
}

func (r AuthenticationRepository) _newSession(tx pgx.Tx, ctx context.Context, now time.Time, user _user.User, fingerprint uuid.UUID, refreshKeyHash user.Password) (domain.Session, refresh.Claims, error) {
	var sessionID string

	err := tx.QueryRow(
		ctx,
		`
			INSERT INTO sessions (user_id, device_fingerprint,  updated_at, created_at)
			VALUES ($1, $2, $3, $3)
			RETURNING id;
			`,
		user.ID,
		fingerprint,
		now,
	).Scan(
		&sessionID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "device_fingerprint_unique" {
			return domain.Session{}, refresh.Claims{}, _domain.ErrDuplicatedFingerprint
		}
		return domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}

	session, err := domain.NewSession(sessionID, fingerprint.String(), nil, now, now)

	if err != nil {
		return domain.Session{}, refresh.Claims{}, err
	}

	var familyID string

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO refresh_tokens (session_id, current_key_hash, expired_at, rotated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING family_id;
		`,
		sessionID,
		refreshKeyHash,
		now.Add(r.refreshKeyDuration),
		now,
	).Scan(&familyID)

	if err != nil {
		return domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}

	refreshClaims, err := refresh.NewClaims(sessionID, familyID, refreshKeyHash)

	if err != nil {
		return domain.Session{}, refresh.Claims{}, err
	}

	return session, refreshClaims, nil
}

func (r AuthenticationRepository) NewSession(ctx context.Context, user _user.User, fingerprint uuid.UUID, refreshKeyHash user.Password) (domain.Session, refresh.Claims, error) {
	tx, err := r.db.Begin(ctx)

	if err != nil {
		return domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}
	defer tx.Rollback(ctx)

	session, claims, err := r._newSession(tx, ctx, time.Now(), user, fingerprint, refreshKeyHash)

	if err != nil {
		return domain.Session{}, refresh.Claims{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}

	return session, claims, nil
}

func (r AuthenticationRepository) Register(ctx context.Context, username _user.Username, email _user.Email, password _user.Password, fingerprint uuid.UUID, refreshKeyHash user.Password) (_user.User, domain.Session, refresh.Claims, error) {
	tx, err := r.db.Begin(ctx)

	if err != nil {
		return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}
	defer tx.Rollback(ctx)

	var userID string
	var now = time.Now()

	err = tx.QueryRow(
		ctx,
		`
			INSERT INTO users (email, username, password_hash, updated_at, created_at)
			VALUES ($1, $2, $3, $4, $4)
			RETURNING id;
			`,
		email,
		username,
		password,
		now,
	).Scan(&userID)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "user_email_unique":
				return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrDuplicatedEmail
			case "user_username_unique":
				return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrDuplicatedUsername
			case "user_email_not_empty":
				return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrInvalidEmail
			case "user_password_hash_not_empty":
				return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrInvalidPassword
			}
		}

		return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}

	user, err := _user.NewUser(
		userID,
		username.String(),
		email.String(),
		password.Expose(),
	)

	if err != nil {
		return _user.User{}, domain.Session{}, refresh.Claims{}, err
	}

	session, refreshClaims, err := r._newSession(tx, ctx, now, user, fingerprint, refreshKeyHash)

	if err != nil {
		return _user.User{}, domain.Session{}, refresh.Claims{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}

	return user,
		session,
		refreshClaims,
		nil
}

func (r AuthenticationRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	tag, err := r.db.Exec(
		ctx,
		`
		DELETE FROM users WHERE id = $1;
		`,
		userID,
	)

	if err != nil {
		return _domain.ErrUnknown
	}

	if tag.RowsAffected() == 0 {
		return _domain.ErrNotFound
	}

	return nil
}

func (r AuthenticationRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	tag, err := r.db.Exec(
		ctx,
		`
			DELETE FROM sessions WHERE id = $1;
		`,
		sessionID,
	)

	if err != nil {
		return _domain.ErrUnknown
	}

	if tag.RowsAffected() == 0 {
		return _domain.ErrNotFound
	}

	return nil
}
