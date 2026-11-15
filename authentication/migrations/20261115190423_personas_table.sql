-- +godwit Up
-- +godwit driver: all
CREATE TABLE IF NOT EXISTS personas (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id      UUID NOT NULL,
    display_name TEXT NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT personas_unique_persona_per_user
        UNIQUE (user_id, display_name),

    CONSTRAINT personas_display_name_not_blank
        CHECK(octet_length(btrim(display_name)) BETWEEN 1 AND 128)
);

-- +godwit Down
-- +godwit driver: all
DROP TABLE IF EXISTS personas;
