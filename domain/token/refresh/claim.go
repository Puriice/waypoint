package refresh

import (
	"uuid"

	"github.com/puriice/waypoint/domain"
	"github.com/puriice/waypoint/domain/errors"
)

type Claims struct {
	SessionID uuid.UUID              `json:"session_id"`
	FamilyID  uuid.UUID              `json:"family_id"`
	Key       domain.SensitiveString `json:"k"`
}

func NewClaims(sessionID, familyID, key string) (Claims, error) {
	sessionUUID, err := uuid.Parse(sessionID)

	if err != nil {
		return Claims{}, errors.NewErrInvalidUUID(sessionID, "session")
	}

	familyUUID, err := uuid.Parse(familyID)

	if err != nil {
		return Claims{}, errors.NewErrInvalidUUID(familyID, "family refresh token")
	}

	return Claims{
		SessionID: sessionUUID,
		FamilyID:  familyUUID,
		Key:       domain.NewSensitiveString(key),
	}, nil
}
