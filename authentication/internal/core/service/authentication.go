package service

import (
	"errors"
	"time"
	"uuid"

	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/domain/cmd"
	"github.com/puriice/waypoint/authentication/internal/core/port"
	"github.com/puriice/waypoint/authentication/internal/core/port/email"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
	"github.com/puriice/waypoint/authentication/internal/infra/config"
	"github.com/puriice/waypoint/domain/app"
	"github.com/puriice/waypoint/domain/serror"
	"github.com/puriice/waypoint/domain/token"
	_user "github.com/puriice/waypoint/domain/user"
)

const _TOKEN_KEY_SIZE = 64
const _DUMMY_INPUT = "this-is-a-dummy-hash"

type repositoryDependency struct {
	authentication repository.Authentication
	loginCounter   repository.LoginCounter
}

type AuthenticationServiceImpl struct {
	repository repositoryDependency
	token      port.AuthenticationTokenizer
	key        port.Key
	hasher     port.Hasher
	email      email.Authentication

	config *config.Config

	dummyHash string
}

func NewAuthenticationService(
	authenticaton repository.Authentication,
	counter repository.LoginCounter,
	token port.AuthenticationTokenizer,
	key port.Key,
	hasher port.Hasher,
	email email.Authentication,
) port.AuthenticationService {
	dummyHash, err := hasher.Hash(_DUMMY_INPUT)

	if err != nil {
		panic(err)
	}

	return AuthenticationServiceImpl{
		repository: repositoryDependency{
			authentication: authenticaton,
			loginCounter:   counter,
		},
		token:  token,
		key:    key,
		hasher: hasher,
		email:  email,

		config: config.Get(),

		dummyHash: dummyHash,
	}
}

func (s AuthenticationServiceImpl) Register(ctx app.Context, command cmd.Register) (_user.User, token.Pair, error) {
	logger := ctx.Logger().With("operation", "register")

	username, err := command.Username()

	if err != nil {
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.InvalidUsername)
	}

	email, err := command.Email()

	if err != nil {
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.InvalidEmail)
	}

	fingerprint, err := command.Fingerprint()

	if err != nil {
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.InvalidDeviceFingerprint)
	}

	password := command.Password()

	if err := password.Apply(s.hasher.Hash); err != nil {
		logger.Error("failed to hash password", "error", err)
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.FailToHash)
	}

	key := s.key.Generate(_TOKEN_KEY_SIZE)
	hashedKey := key

	if err := hashedKey.Apply(s.hasher.Hash); err != nil {
		logger.Error("failed to hash refresh key", "error", err)
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.FailToHash)
	}

	user, session, refreshClaims, err := s.repository.authentication.Register(ctx, username, email, password, fingerprint, hashedKey)

	if err != nil {
		switch {
		case errors.Is(err, _domain.ErrInvalidEmail):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.InvalidEmail)
		case errors.Is(err, _domain.ErrInvalidPassword):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.InvalidPasswordHash)
		case errors.Is(err, _domain.ErrDuplicatedEmail):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.DuplicatedEmail)
		case errors.Is(err, _domain.ErrDuplicatedUsername):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.DuplicatedUsername)
		case errors.Is(err, _domain.ErrDuplicatedFingerprint):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.DuplicatedFingerprint)
		}

		logger.Error("failed to register user", "error", err)
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.UnknownReason)
	}

	logger = logger.With("user_id", user.ID.String(), "session_id", session.ID.String())

	refreshClaims.Key = key

	pair, err := s.token.Sign(ctx, user, session, refreshClaims)

	if err != nil {
		logger.Error("failed to sign tokens; deleting registered user", "error", err)

		go func(ctx app.Context, authRepo repository.Authentication, userID uuid.UUID) {
			if err := authRepo.DeleteUser(ctx, userID); err != nil {
				logger.Error("failed to delete user after signing failure", "error", err)
			}
		}(ctx, s.repository.authentication, user.ID)

		switch {
		case errors.Is(err, _domain.ErrInvalidSigningKey):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.InvalidSigningKey)
		case errors.Is(err, _domain.ErrInvalidTokenClaims):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.InvalidTokenClaims)
		}

		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.FailToSignToken)
	}

	logger.Info("user registered")

	return user, pair, nil
}

func (s AuthenticationServiceImpl) Login(ctx app.Context, command cmd.Login) (_user.User, token.Pair, error) {
	logger := ctx.Logger().With("operation", "login")

	password := command.Password()
	email, err := command.Email()

	if err != nil {
		logger.Info("login rejected", "reason", "invalid email")
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidEmail)
	}

	fingerprint, err := command.Fingerprint()

	if err != nil {
		s.hasher.Verify(password.Expose(), s.dummyHash)
		logger.Info("login rejected", "reason", "invalid device fingerprint")
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	user, err := s.repository.authentication.GetByEmail(ctx, email)

	if err != nil {
		s.hasher.Verify(password.Expose(), s.dummyHash)

		if errors.Is(err, _domain.ErrNotFound) {
			logger.Info("login rejected", "reason", "unknown email")
		} else {
			logger.Error("failed to look up user", "error", err)
		}

		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	logger = logger.With("user_id", user.ID.String())

	if user.IsDuringCooldown(time.Now()) || user.IsDuringQuarantine() {
		s.hasher.Verify(password.Expose(), s.dummyHash)
		logger.Info("login rejected", "reason", "account locked or quarantined")
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	if !s.hasher.Verify(password.Expose(), user.Password.Expose()) {
		logger.Info("login rejected", "reason", "wrong password")
		go s.applyPenalties(ctx, user)
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	key := s.key.Generate(_TOKEN_KEY_SIZE)
	hashedKey := key

	if err := hashedKey.Apply(s.hasher.Hash); err != nil {
		logger.Error("failed to hash refresh key", "error", err)
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	session, claims, err := s.repository.authentication.NewSession(ctx, user, fingerprint, hashedKey)

	if err != nil {
		if errors.Is(err, _domain.ErrDuplicatedFingerprint) {
			logger.Info("login rejected", "reason", "session already exists for device")
		} else {
			logger.Error("failed to create session", "error", err)
		}

		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	logger = logger.With("session_id", session.ID.String())

	claims.Key = key

	pair, err := s.token.Sign(ctx, user, session, claims)

	if err != nil {
		logger.Error("failed to sign tokens; deleting session", "error", err)

		go func(ctx app.Context, authRepo repository.Authentication, sessionID uuid.UUID) {
			if err := authRepo.DeleteSession(ctx, sessionID); err != nil {
				logger.Error("failed to delete session after signing failure", "error", err)
			}
		}(ctx, s.repository.authentication, session.ID)

		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	logger.Info("user logged in")

	return user, pair, nil
}

func (s AuthenticationServiceImpl) applyPenalties(ctx app.Context, user _user.User) {
	logger := ctx.Logger().With("operation", "apply_penalties", "user_id", user.ID.String())

	loginCount, err := s.repository.loginCounter.Count(ctx, user) // start at 1

	if err != nil {
		logger.Error("failed to count failed logins", "error", err)
		return
	}

	if loginCount <= s.config.LoginPenalty.ForgivenWindow {
		return
	}

	penaltyIndex := loginCount - s.config.LoginPenalty.ForgivenWindow - 1
	lastPenaltyIndex := len(s.config.LoginPenalty.Penalties) - 1

	if penaltyIndex > lastPenaltyIndex {
		if !s.config.LoginPenalty.RequireOTP {
			if newUser, success := s.QuarantineAccount(ctx, user); success {
				return
			} else {
				user = newUser
			}
		}
		penaltyIndex = lastPenaltyIndex
	}

	lockedUntil := time.Now().Add(time.Duration(s.config.LoginPenalty.Penalties[penaltyIndex]))

	_, err = s.repository.authentication.Lock(ctx, user, lockedUntil)

	if err != nil {
		logger.Error("failed to lock account", "failed_logins", loginCount, "error", err)
		return
	}

	logger.Info("account locked", "failed_logins", loginCount, "locked_until", lockedUntil)
}

func (s AuthenticationServiceImpl) QuarantineAccount(ctx app.Context, user _user.User) (_user.User, bool) {
	logger := ctx.Logger().With("operation", "quarantine_account", "user_id", user.ID.String())

	unlockToken := s.key.Generate(_TOKEN_KEY_SIZE)
	unlockTokenHash := unlockToken

	if err := unlockTokenHash.Apply(s.hasher.Hash); err != nil {
		logger.Error("failed to hash unlock token", "error", err)
		return user, false
	}

	newUser, err := s.repository.authentication.Block(ctx, user, unlockTokenHash)

	if err != nil {
		logger.Error("failed to block account", "error", err)
		return user, false
	}

	user = newUser

	err = s.email.SendUnlockLink(ctx, user.Email, unlockToken)

	if err != nil {
		logger.Error("failed to send unlock link; unblocking account", "error", err)

		newUser, err := s.repository.authentication.Unblock(ctx, user)

		if err != nil {
			logger.Error("failed to unblock account after email failure", "error", err)
			return user, false
		}

		return newUser, false
	}

	logger.Info("account quarantined")

	return user, true
}
