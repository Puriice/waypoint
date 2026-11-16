package postgres

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/token/refresh"
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

func (r AuthenticationRepository) Login(ctx context.Context, username _user.Username, email _user.Email, password _user.Password, fingerprint uuid.UUID, refreshKeyHash string) (_user.User, domain.Session, refresh.Claims, error) {
	panic("unimplemented")
}

func (r AuthenticationRepository) Register(ctx context.Context, username _user.Username, email _user.Email, password _user.Password, fingerprint uuid.UUID, refreshKeyHash string) (_user.User, domain.Session, refresh.Claims, error) {
	tx, err := r.db.Begin(ctx)
	defer tx.Rollback(ctx)

	if err != nil {
		panic("unimplemented")
	}

	var userID string
	var now = time.Now()

	err = tx.QueryRow(
		ctx,
		`
			INSERT INTO users (email, username, password_hash, updated_at, created_at)
			VALUES ($1, $2, $3, $4, $4)
			RETURNING id;
			`,
		email.String(),
		username.String(),
		password.Expose(),
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

	var sessionID string

	err = tx.QueryRow(
		ctx,
		`
			INSERT INTO sessions (user_id, device_fingerprint,  updated_at, created_at)
			VALUES ($1, $2, $3, $3)
			RETURNING id, expired_at;
			`,
		userID,
		fingerprint,
		now,
	).Scan(
		&sessionID,
	)

	if err != nil {
		return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}

	session, err := domain.NewSession(sessionID, fingerprint.String(), nil, now, now)

	if err != nil {
		return _user.User{}, domain.Session{}, refresh.Claims{}, err
	}

	var familyID string
	var expiredAt time.Time

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO refresh_tokens (session_id, current_key_hash, expired_at, rotated_at)
		VALUES ($1, $2, $3, $4)
		`,
		sessionID,
		refreshKeyHash,
		now.Add(r.refreshKeyDuration),
		now,
	).Scan(&familyID, &expiredAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "device_fingerprint_unique" {
			return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrDuplicatedFingerprint
		}
		return _user.User{}, domain.Session{}, refresh.Claims{}, _domain.ErrUnknown
	}

	refreshClaims, err := refresh.NewClaims(sessionID, familyID, refreshKeyHash)

	if err != nil {
		return _user.User{}, domain.Session{}, refresh.Claims{}, err
	}

	tx.Commit(ctx)

	return user,
		session,
		refreshClaims,
		nil
}

func (r AuthenticationRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	tag, err := r.db.Exec(
		ctx,
		`
		DELETE FROM users WHERE id = $1;
		`,
		userID.String(),
	)

	if err != nil {
		return _domain.ErrUnknown
	}

	if tag.RowsAffected() == 0 {
		return _domain.ErrUserNotFound
	}

	return nil
}
