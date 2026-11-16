package user_test

import (
	"testing"

	"github.com/puriice/waypoint/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsernameScan(t *testing.T) {
	var u user.Username

	require.NoError(t, u.Scan("alice"))
	assert.Equal(t, user.Username("alice"), u)
}

func TestUsernameScanRejectsNonString(t *testing.T) {
	var u user.Username

	require.Error(t, u.Scan(42))
}

func TestEmailScan(t *testing.T) {
	var e user.Email

	require.NoError(t, e.Scan("alice@example.com"))
	assert.Equal(t, user.Email("alice@example.com"), e)
}

func TestEmailScanRejectsInvalidEmail(t *testing.T) {
	e := user.Email("before@example.com")

	require.ErrorIs(t, e.Scan("not-an-email"), user.ErrInvalidEmail)
	assert.Equal(t, user.Email("before@example.com"), e, "failed scan must not overwrite the value")
}
