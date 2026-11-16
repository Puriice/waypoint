-- +godwit Up
-- +godwit driver: all
CREATE TABLE IF NOT EXISTS users (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    version             INT NOT NULL DEFAULT 0,
    email               TEXT NOT NULL,
    username			TEXT NOT NULL,
    password_hash       TEXT NOT NULL,
    email_verified_at   TIMESTAMPTZ,
    deactivated_at      TIMESTAMPTZ,
    locked_until        TIMESTAMPTZ,
    quarantine_at       TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT user_email_unique
    	UNIQUE(email),
    CONSTRAINT user_username_unique
    	UNIQUE(username),
    CONSTRAINT user_email_not_empty
        CHECK (octet_length(btrim(email)) > 0),
    CONSTRAINT user_password_hash_not_empty
        CHECK (octet_length(btrim(password_hash)) > 32)
);

-- +godwit Down
-- +godwit driver: all
DROP TABLE IF EXISTS users;
