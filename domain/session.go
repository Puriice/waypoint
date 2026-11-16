package domain

import (
	"time"
	"uuid"

	"github.com/puriice/waypoint/domain/errors"
)

type Session struct {
	ID                uuid.UUID
	DeviceFingerprint uuid.UUID
	MFAVerifiedAt     *time.Time
	UpdatedAt         time.Time
	CreatedAt         time.Time
}

func NewSession(id, fingerprint string, mfaVerifiedAt *time.Time, updatedAt, createdAt time.Time) (Session, error) {
	sessionID, err := uuid.Parse(id)

	if err != nil {
		return Session{}, errors.NewErrInvalidUUID(id, "session")
	}

	deviceFingerprint, err := uuid.Parse(fingerprint)

	if err != nil {
		return Session{}, errors.NewErrInvalidUUID(id, "fingerprint")
	}

	return Session{
		ID:                sessionID,
		DeviceFingerprint: deviceFingerprint,
		MFAVerifiedAt:     mfaVerifiedAt,
		UpdatedAt:         updatedAt,
		CreatedAt:         createdAt,
	}, nil
}
