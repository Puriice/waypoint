package user_test

import (
	"encoding/json"
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

func TestPasswordJSONIsPlainString(t *testing.T) {
	data, err := json.Marshal(struct {
		Key user.Password `json:"k"`
	}{user.NewPassword("secret")})
	require.NoError(t, err)
	assert.JSONEq(t, `{"k":"secret"}`, string(data))

	var got struct {
		Key user.Password `json:"k"`
	}
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, "secret", got.Key.Expose())
}
