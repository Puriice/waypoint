package service_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/puriice/waypoint/authentication/internal/adapter/repository/inmemory"
	_domain "github.com/puriice/waypoint/authentication/internal/core/domain"
	"github.com/puriice/waypoint/authentication/internal/core/domain/cmd"
	"github.com/puriice/waypoint/authentication/internal/core/port"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
	"github.com/puriice/waypoint/authentication/internal/core/service"
	"github.com/puriice/waypoint/domain/app"
	"github.com/puriice/waypoint/domain/serror"
	_user "github.com/puriice/waypoint/domain/user"
)

func TestLoginUnknownEmailReturnsInvalidCredentialAfterVerifying(t *testing.T) {
	hasher := &recordingHasher{}
	svc := service.NewAuthenticationService(inmemory.NewAuthentication(time.Hour), nil, nil, hasher)

	_, _, err := svc.Login(newContext(), cmd.NewLogin("nobody@example.com", "secret", uuid.NewV4().String()))

	require.ErrorIs(t, err, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential))
	assert.Equal(t, 1, hasher.verifyCalls, "unknown email must still cost one password verification")
}

func TestLoginRepositoryFailureReturnsInvalidCredential(t *testing.T) {
	hasher := &recordingHasher{}
	svc := service.NewAuthenticationService(failingRepository{}, nil, nil, hasher)

	_, _, err := svc.Login(newContext(), cmd.NewLogin("someone@example.com", "secret", uuid.NewV4().String()))

	require.ErrorIs(t, err, serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential))
	assert.Equal(t, 1, hasher.verifyCalls, "repository failure must still cost one password verification")
}

func TestLoginInvalidInputReturnsInvalidCredential(t *testing.T) {
	svc := service.NewAuthenticationService(inmemory.NewAuthentication(time.Hour), nil, nil, &recordingHasher{})
	want := serror.NewServiceError(serror.Authentication, serror.Login, serror.InvalidCredential)

	_, _, err := svc.Login(newContext(), cmd.NewLogin("not-an-email", "secret", uuid.NewV4().String()))
	require.ErrorIs(t, err, want, "invalid email")

	_, _, err = svc.Login(newContext(), cmd.NewLogin("someone@example.com", "secret", "not-a-uuid"))
	require.ErrorIs(t, err, want, "invalid fingerprint")
}

func newContext() app.Context {
	return app.Context{Context: context.Background()}
}

type recordingHasher struct {
	verifyCalls int
}

var _ port.Hasher = (*recordingHasher)(nil)

func (h *recordingHasher) Hash(input string) (string, error) { return "hashed:" + input, nil }

func (h *recordingHasher) Verify(input, hash string) bool {
	h.verifyCalls++
	return hash == "hashed:"+input
}

// failingRepository simulates an unavailable database.
type failingRepository struct {
	repository.Authentication
}

func (failingRepository) GetByEmail(context.Context, _user.Email) (_user.User, error) {
	return _user.User{}, _domain.ErrUnknown
}
