package token

import (
	"github.com/puriice/waypoint/domain/token/access"
	"github.com/puriice/waypoint/domain/token/refresh"
)

type Pair struct {
	AccessToken  access.Token
	RefreshToken refresh.Token
}

func NewPair(access access.Token, refresh refresh.Token) Pair {
	return Pair{
		AccessToken:  access,
		RefreshToken: refresh,
	}
}
