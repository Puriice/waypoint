package service

import (
	"context"
	"errors"
	"uuid"

	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/domain/cmd"
	"github.com/puriice/waypoint/authentication/internal/core/port"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
	"github.com/puriice/waypoint/domain/app"
	"github.com/puriice/waypoint/domain/serror"
	"github.com/puriice/waypoint/domain/token"
	_user "github.com/puriice/waypoint/domain/user"
)

const REFRESH_KEY_SIZE = 64
const dummyHash = "this-is-a-dummy-hash"

type repositoryDependency struct {
	authentication repository.Authentication
}

type AuthenticationServiceImpl struct {
	repository repositoryDependency
	token      port.AuthenticationTokenizer
	key        port.Key
	hasher     port.Hasher
}

func NewAuthenticationService(
	authenticaton repository.Authentication,
	token port.AuthenticationTokenizer,
	key port.Key,
	hasher port.Hasher,
) port.AuthenticationService {
	return AuthenticationServiceImpl{
		repository: repositoryDependency{
			authentication: authenticaton,
		},
		token:  token,
		key:    key,
		hasher: hasher,
	}
}

func (s AuthenticationServiceImpl) Register(ctx app.Context, command cmd.Register) (_user.User, token.Pair, error) {
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
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.FailToHash)
	}

	key := s.key.Generate(REFRESH_KEY_SIZE)
	hashedKey := key

	if err := hashedKey.Apply(s.hasher.Hash); err != nil {
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

		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Register, serror.UnknownReason)
	}

	refreshClaims.Key = key

	pair, err := s.token.Sign(ctx, user, session, refreshClaims)

	if err != nil {
		go func(ctx context.Context, authRepo repository.Authentication, userID uuid.UUID) {
			authRepo.DeleteUser(ctx, userID)
		}(ctx, s.repository.authentication, user.ID)

		switch {
		case errors.Is(err, _domain.ErrInvalidSigningKey):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.InvalidSigningKey)
		case errors.Is(err, _domain.ErrInvalidTokenClaims):
			return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.InvalidTokenClaims)
		}

		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.System, serror.General, serror.FailToSignToken)
	}

	return user, pair, nil
}

func (s AuthenticationServiceImpl) Login(ctx app.Context, command cmd.Login) (_user.User, token.Pair, error) {
	email, err := command.Email()

	if err != nil {
		s.hasher.Verify("password", dummyHash)
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	fingerprint, err := command.Fingerprint()

	if err != nil {
		s.hasher.Verify("password", dummyHash)
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	password := command.Password()

	user, err := s.repository.authentication.GetByEmail(ctx, email)

	if !s.hasher.Verify(password.Expose(), user.Password.Expose()) {
		go s.applyPenalties(ctx, user)
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	key := s.key.Generate(REFRESH_KEY_SIZE)
	hashedKey := key

	if err := hashedKey.Apply(s.hasher.Hash); err != nil {
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	session, claims, err := s.repository.authentication.NewSession(ctx, user, fingerprint, hashedKey)

	if err != nil {
		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	pair, err := s.token.Sign(ctx, user, session, claims)

	if err != nil {
		go func(ctx context.Context, authRepo repository.Authentication, sessionID uuid.UUID) {
			authRepo.DeleteSession(ctx, sessionID)
		}(ctx, s.repository.authentication, session.ID)

		return _user.User{}, token.Pair{}, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)
	}

	return user, pair, nil
}

func (s AuthenticationServiceImpl) applyPenalties(ctx app.Context, user _user.User) {

}
