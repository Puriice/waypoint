package email

import (
	"context"

	"github.com/puriice/waypoint/domain/user"
)

type Authentication interface {
	SendUnlockLink(ctx context.Context, to user.Email, token user.Password) error
}
