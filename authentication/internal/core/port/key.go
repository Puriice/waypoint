package port

import "github.com/puriice/waypoint/domain"

type Key interface {
	Generate(size int) domain.SensitiveString
}
