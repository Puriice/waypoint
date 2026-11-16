package inmemory_test

import (
	"testing"
	"time"

	adapter "github.com/puriice/waypoint/authentication/internal/adapter/repository"
	"github.com/puriice/waypoint/authentication/internal/adapter/repository/inmemory"
	"github.com/puriice/waypoint/authentication/internal/core/port/repository"
)

func TestAuthenticationRepository(t *testing.T) {
	adapter.RunAuthenticationSuite(t, func(t *testing.T) repository.Authentication {
		return inmemory.NewAuthentication(time.Hour)
	})
}
