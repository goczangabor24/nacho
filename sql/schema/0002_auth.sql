-- +goose Up
ALTER TABLE users
ADD COLUMN display_name TEXT NOT NULL DEFAULT '',
ADD COLUMN role TEXT NOT NULL DEFAULT 'waiter',
ADD COLUMN disabled_at TIMESTAMPTZ,
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE users
ADD CONSTRAINT users_role_check
check(role in ('admin', 'manager', 'waiter'));

CREATE UNIQUE INDEX users_username_lower_idx ON users (lower(username));

ALTER TABLE SESSIONS
ADD COLUMN last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- +goose Down
ALTER TABLE SESSIONS DROP COLUMN last_seen_at;
DROP INDEX users_username_lower_idx;
ALTER TABLE USERS DROP CONSTRAINT users_role_check;
ALTER TABLE USERS
DROP COLUMN updated_at,
DROP COLUMN disabled_at,
DROP COLUMN role,
DROP COLUMN display_name;