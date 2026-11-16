package user_test

import (
	"testing"

	"github.com/puriice/waypoint/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidEmail(t *testing.T) {
	email := "test@example.com"

	parsedEmail, err := user.NewEmail(email)
	require.NoError(t, err)

	assert.Equal(t, email, string(parsedEmail))
}

func TestInvalidEmail(t *testing.T) {
	t.Parallel()

	invalidEmails := []string{
		"plainaddress",
		"@missingusername.com",
		"username@.com",
		"username@com",
		"username@domain..com",
		"username@domain.com.",
		".username@domain.com",
		"username.@domain.com",
		"username@domain.c",
		"username@-domain.com",
		"username@domain-.com",
		"user name@domain.com",
		"username@domain,com",
		"username@@domain.com",
		"username@domain.com#",
		"username@domain..com.",
		"username@[123.123.123.123]extra",
		"username@domain..",
		"username@",
		"#@%^%#$@#$@#.com",
		"@example.com",
		"Joe Smith <email@example.com>",
		"email.example.com",
		"email@example@example.com",
		".email@example.com",
		"email.@example.com",
		"email..email@example.com",
		"あいうえお@example.com",
		"email@example.com (Joe Smith)",
		"email@example",
		"email@-example.com",
		"email@example.web",
		"email@111.222.333.44444",
		"email@example..com",
		"Abc..123@example.com",
		"\"(),:;<>[\\]@example.com",
		"just”not”right@example.com",
		"this\\ is\"really\"not\allowed@example.com",
	}

	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			t.Parallel()
			_, err := user.NewEmail(email)

			assert.ErrorIs(t, err, user.ErrInvalidEmail, email)
		})
	}
}
