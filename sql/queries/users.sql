-- name: CreateUser :one
insert into users (username, display_name, role, password_hash)
values (sqlc.arg(username), sqlc.arg(display_name), sqlc.arg(role), sqlc.arg(password_hash))
returning id;

-- name: GetUserByUsername :one
select id, username, display_name, role, password_hash, disabled_at
from users
where lower(username) = lower(sqlc.arg(username));

-- name: ListUsers :many
select id, username, display_name, role, disabled_at, created_at
from users
order by lower(username);

-- name: SetUserRole :exec
update users
set role = sqlc.arg(role), updated_at = now()
where id = sqlc.arg(id);

-- name: SetUserPassword :exec
update users
set password_hash = sqlc.arg(password_hash), updated_at = now()
where id = sqlc.arg(id);

-- name: SetUserDisabled :exec
update users
set disabled_at = case when sqlc.arg(disabled)::boolean
                       then coalesce(disabled_at, now())
                       else null end,
    updated_at = now()
where id = sqlc.arg(id);
