package config

import "github.com/puriice/waypoint/domain/unit"

type LoginPenalty struct {
	ForgivenWindow int            `json:"forgiven_window"`
	Penalties      []unit.Seconds `json:"penalties"`
	RequireOTP     bool           `json:"otp_required"`
}
