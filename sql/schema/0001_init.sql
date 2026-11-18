-- +goose Up
create table users (
    id            bigint generated always as identity primary key,
    username      text        not null unique,
    password_hash text        not null,
    created_at    timestamptz not null default now()
);

create table sessions (
    token_hash bytea       primary key,
    user_id    bigint      not null references users (id) on delete cascade,
    expires_at timestamptz not null,
    created_at timestamptz not null default now()
);

create index sessions_user_id_idx on sessions (user_id);
create index sessions_expires_at_idx on sessions (expires_at);

-- +goose Down
drop table sessions;
drop table users;
