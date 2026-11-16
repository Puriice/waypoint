-- +godwit Up
-- +godwit driver: all
CREATE TABLE IF NOT EXISTS sessions (
	id					UUID PRIMARY KEY DEFAULT uuidv7(),
	user_id 			UUID NOT NULL,
	device_fingerprint	UUID NOT NULL,
	mfa_verified_at		TIMESTAMPTZ,
	updated_at 			TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_at 			TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
	CONSTRAINT device_fingerprint_unique
		UNIQUE(user_id, device_fingerprint)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
	session_id			UUID PRIMARY KEY,
	family_id			UUID NOT NULL DEFAULT uuidv7(),
	current_key_hash 	TEXT NOT NULL,
	previous_key_hash 	TEXT,
	expired_at			TIMESTAMPTZ NOT NULL,
	rotated_at			TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	revoked_at			TIMESTAMPTZ,
	FOREIGN KEY (session_id) REFERENCES sessions(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- +godwit Down
-- +godwit driver: all
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS sessions;
