-- name: CreateAgentHandoff :one
INSERT INTO agent_handoffs (
  id, team_id, issue_id, from_actor_type, from_bot_id, from_user_id,
  to_bot_id, trigger_comment_id, source_session_id, target_session_id,
  status, metadata
) VALUES (
  lower(hex(randomblob(4))) || '-' ||
  lower(hex(randomblob(2))) || '-' ||
  '4' || substr(lower(hex(randomblob(2))), 2) || '-' ||
  substr('89ab', abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))), 2) || '-' ||
  lower(hex(randomblob(6))),
  sqlc.arg(team_id),
  sqlc.narg(issue_id),
  sqlc.arg(from_actor_type),
  sqlc.narg(from_bot_id),
  sqlc.narg(from_user_id),
  sqlc.arg(to_bot_id),
  sqlc.narg(trigger_comment_id),
  sqlc.narg(source_session_id),
  sqlc.narg(target_session_id),
  sqlc.arg(status),
  sqlc.arg(metadata)
)
RETURNING *;

-- name: GetAgentHandoff :one
SELECT * FROM agent_handoffs WHERE id = sqlc.arg(id);

-- name: ListPendingHandoffsToBotForIssue :many
SELECT * FROM agent_handoffs
WHERE to_bot_id = sqlc.arg(to_bot_id)
  AND issue_id = sqlc.arg(issue_id)
  AND status IN ('pending', 'dispatched', 'running')
ORDER BY created_at DESC;

-- name: ListPendingHandoffsToBot :many
SELECT * FROM agent_handoffs
WHERE to_bot_id = sqlc.arg(to_bot_id)
  AND status IN ('pending', 'dispatched', 'running')
ORDER BY created_at ASC;

-- name: ListPendingReturnsForBotInIssue :many
SELECT * FROM agent_handoffs
WHERE from_bot_id = sqlc.arg(from_bot_id)
  AND issue_id = sqlc.arg(issue_id)
  AND status = 'completed'
  AND return_handoff_id IS NULL
ORDER BY completed_at ASC;

-- name: ListHandoffsByIssue :many
SELECT * FROM agent_handoffs
WHERE issue_id = sqlc.arg(issue_id)
ORDER BY created_at ASC;

-- name: MarkHandoffDispatched :one
UPDATE agent_handoffs SET
  status = 'dispatched',
  target_session_id = COALESCE(sqlc.narg(target_session_id), target_session_id),
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND status = 'pending'
RETURNING *;

-- name: MarkHandoffRunning :one
UPDATE agent_handoffs SET
  status = 'running',
  target_session_id = COALESCE(sqlc.narg(target_session_id), target_session_id),
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND status IN ('pending', 'dispatched')
RETURNING *;

-- name: CompleteHandoff :one
UPDATE agent_handoffs SET
  status = 'completed',
  result_comment_id = sqlc.narg(result_comment_id),
  completed_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND status IN ('pending', 'dispatched', 'running')
RETURNING *;

-- name: FailHandoff :one
UPDATE agent_handoffs SET
  status = 'failed',
  failure_reason = sqlc.arg(failure_reason),
  completed_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND status IN ('pending', 'dispatched', 'running')
RETURNING *;

-- name: SetHandoffReturn :one
UPDATE agent_handoffs SET
  return_handoff_id = sqlc.arg(return_handoff_id),
  status = 'returned',
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND status IN ('completed', 'returned')
RETURNING *;

-- name: CancelHandoff :one
UPDATE agent_handoffs SET
  status = 'cancelled',
  failure_reason = sqlc.arg(failure_reason),
  completed_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND status IN ('pending', 'dispatched', 'running')
RETURNING *;
