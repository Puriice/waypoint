package config

import "github.com/puriice/waypoint/domain/unit"

type ttl struct {
	AccessToken  unit.Seconds `json:"access_token"`
	RefreshToken unit.Seconds `json:"refresh_token"`
}

type jwtPolicy struct {
	Issuer string `json:"issuer"`
	TTL    ttl    `json:"ttl_in_seconds"`
}
