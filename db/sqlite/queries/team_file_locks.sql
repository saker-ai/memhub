-- name: AcquireTeamFileLock :one
INSERT INTO team_file_locks (
  id, team_id, path, scope, owner_type, owner_bot_id, owner_user_id,
  issue_id, session_id, handoff_id, expires_at, metadata
) VALUES (
  lower(hex(randomblob(4))) || '-' ||
  lower(hex(randomblob(2))) || '-' ||
  '4' || substr(lower(hex(randomblob(2))), 2) || '-' ||
  substr('89ab', abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))), 2) || '-' ||
  lower(hex(randomblob(6))),
  sqlc.arg(team_id),
  sqlc.arg(path),
  sqlc.arg(scope),
  sqlc.arg(owner_type),
  sqlc.narg(owner_bot_id),
  sqlc.narg(owner_user_id),
  sqlc.narg(issue_id),
  sqlc.narg(session_id),
  sqlc.narg(handoff_id),
  sqlc.arg(expires_at),
  sqlc.arg(metadata)
)
ON CONFLICT (team_id, path, scope) DO NOTHING
RETURNING *;

-- name: GetTeamFileLock :one
SELECT * FROM team_file_locks
WHERE team_id = sqlc.arg(team_id) AND path = sqlc.arg(path) AND scope = sqlc.arg(scope);

-- name: GetTeamFileLockByID :one
SELECT * FROM team_file_locks WHERE id = sqlc.arg(id);

-- name: ListTeamFileLocks :many
SELECT * FROM team_file_locks
WHERE team_id = sqlc.arg(team_id)
ORDER BY acquired_at DESC;

-- name: ListActiveTeamFileLocks :many
SELECT * FROM team_file_locks
WHERE team_id = sqlc.arg(team_id)
  AND expires_at > CURRENT_TIMESTAMP
ORDER BY acquired_at DESC;

-- name: RefreshTeamFileLock :one
UPDATE team_file_locks SET
  refreshed_at = CURRENT_TIMESTAMP,
  expires_at = sqlc.arg(expires_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ReleaseTeamFileLock :exec
DELETE FROM team_file_locks WHERE id = sqlc.arg(id);

-- name: ReleaseExpiredTeamFileLocks :exec
DELETE FROM team_file_locks
WHERE expires_at <= CURRENT_TIMESTAMP;
