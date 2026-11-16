package port

import (
	"github.com/puriice/waypoint/domain/user"
)

type Key interface {
	Generate(size int) user.Password
}
