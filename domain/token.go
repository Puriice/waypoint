package domain

import "github.com/golang-jwt/jwt/v5"

type AccessToken struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
}

type RefreshToken struct {
	jwt.RegisteredClaims
	FamilyID  string   `json:"family_id"`
	SessionID string   `json:"session_id"`
	Personas  []string `json:"personas"`
}
