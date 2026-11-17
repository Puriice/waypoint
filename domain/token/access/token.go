package access

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/user"
)

type Token struct {
	jwt.RegisteredClaims
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
	signed    domain.SensitiveString `json:"-"`
}

// NewToken builds an access token. signed is the encoded JWT the claims came
// from, or an empty value for a token that has not been signed yet.
func NewToken(claims jwt.RegisteredClaims, user user.User, session domain.Session, signed domain.SensitiveString) Token {
	claims.Subject = user.ID.String()

	return Token{
		RegisteredClaims: claims,
		Email:            user.Email.String(),
		SessionID:        session.ID.String(),
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
