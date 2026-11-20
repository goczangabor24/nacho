-- name: CreateSession :exec
insert into sessions (token_hash, user_id, expires_at, last_seen_at)
values (
    sqlc.arg(token_hash),
    sqlc.arg(user_id),
    sqlc.arg(expires_at),
    sqlc.arg(last_seen_at)
);

-- name: GetSessionByToken :one
select s.token_hash, s.user_id, s.created_at, s.expires_at, s.last_seen_at,
       u.display_name, u.role, u.disabled_at
from sessions s
join users u on u.id = s.user_id
where s.token_hash = sqlc.arg(token_hash);

-- name: TouchSession :exec
update sessions
set expires_at = sqlc.arg(expires_at), last_seen_at = sqlc.arg(last_seen_at)
where token_hash = sqlc.arg(token_hash);

-- name: DeleteSession :exec
delete from sessions where token_hash = sqlc.arg(token_hash);

-- name: DeleteSessionsForUser :exec
delete from sessions where user_id = sqlc.arg(user_id);

-- name: DeleteExpiredSessions :execrows
delete from sessions
where expires_at <= sqlc.arg(now) or created_at <= sqlc.arg(cutoff);
