package refresh

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	jwt.RegisteredClaims
	Claims
	signed string
}

func NewToken(registerClaims jwt.RegisteredClaims, claims Claims) Token {
	return Token{
		RegisteredClaims: registerClaims,
		Claims:           claims,
	}
}

func (t *Token) Sign(key *rsa.PrivateKey) error {
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, t).SignedString(key)

	if err != nil {
		return err
	}

	t.signed = signed

	return nil
}

func (t Token) SignedToken() string {
	return t.signed
}
