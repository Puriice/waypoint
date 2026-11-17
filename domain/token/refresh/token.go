package refresh

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
	"github.com/puriice/waypoint/domain"
)

type Token struct {
	jwt.RegisteredClaims
	Claims
	signed domain.SensitiveString `json:"-"`
}

// NewToken builds a refresh token. signed is the encoded JWT the claims came
// from, or an empty value for a token that has not been signed yet.
func NewToken(registerClaims jwt.RegisteredClaims, claims Claims, signed domain.SensitiveString) Token {
	return Token{
		RegisteredClaims: registerClaims,
		Claims:           claims,
		signed:           signed,
	}
}

func (t *Token) Sign(key *rsa.PrivateKey) error {
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, t).SignedString(key)

	if err != nil {
		return err
	}

	t.signed = domain.NewSensitiveString(signed)

	return nil
}

func (t Token) SignedToken() domain.SensitiveString {
	return t.signed
}
