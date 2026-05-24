-- 0006_agent_teams
-- Agent Team collaboration: teams, members, issues, comments, handoffs and shared FS file locks.

CREATE TABLE IF NOT EXISTS agent_teams (
  id TEXT PRIMARY KEY,
  owner_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  shared_dir_name TEXT NOT NULL DEFAULT '',
  instructions TEXT NOT NULL DEFAULT '',
  metadata TEXT NOT NULL DEFAULT '{}',
  archived_at TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT agent_teams_owner_name_unique UNIQUE (owner_user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_agent_teams_owner ON agent_teams(owner_user_id);

CREATE TABLE IF NOT EXISTS agent_team_members (
  id TEXT PRIMARY KEY,
  team_id TEXT NOT NULL REFERENCES agent_teams(id) ON DELETE CASCADE,
  member_type TEXT NOT NULL,
  bot_id TEXT REFERENCES bots(id) ON DELETE CASCADE,
  user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT '',
  instructions TEXT NOT NULL DEFAULT '',
  metadata TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT agent_team_members_type_check CHECK (member_type IN ('bot', 'user')),
  CONSTRAINT agent_team_members_target_check CHECK (
    (member_type = 'bot' AND bot_id IS NOT NULL AND user_id IS NULL)
    OR (member_type = 'user' AND user_id IS NOT NULL AND bot_id IS NULL)
  ),
  CONSTRAINT agent_team_members_bot_unique UNIQUE (team_id, bot_id),
  CONSTRAINT agent_team_members_user_unique UNIQUE (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_team_members_team ON agent_team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_agent_team_members_bot ON agent_team_members(bot_id);
CREATE INDEX IF NOT EXISTS idx_agent_team_members_user ON agent_team_members(user_id);

CREATE TABLE IF NOT EXISTS team_issues (
  id TEXT PRIMARY KEY,
  team_id TEXT NOT NULL REFERENCES agent_teams(id) ON DELETE CASCADE,
  number INTEGER NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'todo',
  assignee_type TEXT,
  assignee_bot_id TEXT REFERENCES bots(id) ON DELETE SET NULL,
  assignee_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_by_type TEXT NOT NULL,
  created_by_bot_id TEXT REFERENCES bots(id) ON DELETE SET NULL,
  created_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
  parent_issue_id TEXT REFERENCES team_issues(id) ON DELETE SET NULL,
  metadata TEXT NOT NULL DEFAULT '{}',
  closed_at TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT team_issues_status_check CHECK (status IN ('backlog', 'todo', 'in_progress', 'blocked', 'review', 'done', 'cancelled')),
  CONSTRAINT team_issues_assignee_type_check CHECK (assignee_type IS NULL OR assignee_type IN ('bot', 'user')),
  CONSTRAINT team_issues_created_by_type_check CHECK (created_by_type IN ('bot', 'user', 'system')),
  CONSTRAINT team_issues_team_number_unique UNIQUE (team_id, number)
);

CREATE INDEX IF NOT EXISTS idx_team_issues_team_status ON team_issues(team_id, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_team_issues_assignee_bot ON team_issues(assignee_bot_id);
CREATE INDEX IF NOT EXISTS idx_team_issues_assignee_user ON team_issues(assignee_user_id);
CREATE INDEX IF NOT EXISTS idx_team_issues_parent ON team_issues(parent_issue_id);

CREATE TABLE IF NOT EXISTS team_issue_comments (
  id TEXT PRIMARY KEY,
  issue_id TEXT NOT NULL REFERENCES team_issues(id) ON DELETE CASCADE,
  team_id TEXT NOT NULL REFERENCES agent_teams(id) ON DELETE CASCADE,
  parent_comment_id TEXT REFERENCES team_issue_comments(id) ON DELETE SET NULL,
  author_type TEXT NOT NULL,
  author_bot_id TEXT REFERENCES bots(id) ON DELETE SET NULL,
  author_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
  content TEXT NOT NULL DEFAULT '',
  metadata TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT team_issue_comments_author_type_check CHECK (author_type IN ('bot', 'user', 'system'))
);

CREATE INDEX IF NOT EXISTS idx_team_issue_comments_issue ON team_issue_comments(issue_id, created_at);
CREATE INDEX IF NOT EXISTS idx_team_issue_comments_team ON team_issue_comments(team_id);

CREATE TABLE IF NOT EXISTS agent_handoffs (
  id TEXT PRIMARY KEY,
  team_id TEXT NOT NULL REFERENCES agent_teams(id) ON DELETE CASCADE,
  issue_id TEXT REFERENCES team_issues(id) ON DELETE CASCADE,
  from_actor_type TEXT NOT NULL,
  from_bot_id TEXT REFERENCES bots(id) ON DELETE SET NULL,
  from_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
  to_bot_id TEXT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  trigger_comment_id TEXT REFERENCES team_issue_comments(id) ON DELETE SET NULL,
  source_session_id TEXT REFERENCES bot_sessions(id) ON DELETE SET NULL,
  target_session_id TEXT REFERENCES bot_sessions(id) ON DELETE SET NULL,
  result_comment_id TEXT REFERENCES team_issue_comments(id) ON DELETE SET NULL,
  return_handoff_id TEXT REFERENCES agent_handoffs(id) ON DELETE SET NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  failure_reason TEXT NOT NULL DEFAULT '',
  metadata TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at TEXT,
  CONSTRAINT agent_handoffs_from_actor_type_check CHECK (from_actor_type IN ('bot', 'user', 'system')),
  CONSTRAINT agent_handoffs_status_check CHECK (status IN ('pending', 'dispatched', 'running', 'completed', 'failed', 'cancelled', 'returned'))
);

CREATE INDEX IF NOT EXISTS idx_agent_handoffs_team ON agent_handoffs(team_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_handoffs_issue ON agent_handoffs(issue_id);
CREATE INDEX IF NOT EXISTS idx_agent_handoffs_to_bot ON agent_handoffs(to_bot_id, status);
CREATE INDEX IF NOT EXISTS idx_agent_handoffs_from_bot ON agent_handoffs(from_bot_id, status);

CREATE TABLE IF NOT EXISTS team_file_locks (
  id TEXT PRIMARY KEY,
  team_id TEXT NOT NULL REFERENCES agent_teams(id) ON DELETE CASCADE,
  path TEXT NOT NULL,
  scope TEXT NOT NULL DEFAULT 'file',
  owner_type TEXT NOT NULL,
  owner_bot_id TEXT REFERENCES bots(id) ON DELETE CASCADE,
  owner_user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
  issue_id TEXT REFERENCES team_issues(id) ON DELETE SET NULL,
  session_id TEXT REFERENCES bot_sessions(id) ON DELETE SET NULL,
  handoff_id TEXT REFERENCES agent_handoffs(id) ON DELETE SET NULL,
  acquired_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  refreshed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at TEXT NOT NULL,
  metadata TEXT NOT NULL DEFAULT '{}',
  CONSTRAINT team_file_locks_owner_type_check CHECK (owner_type IN ('bot', 'user', 'system')),
  CONSTRAINT team_file_locks_scope_check CHECK (scope IN ('file', 'prefix')),
  CONSTRAINT team_file_locks_team_path_unique UNIQUE (team_id, path, scope)
);

CREATE INDEX IF NOT EXISTS idx_team_file_locks_team_expires ON team_file_locks(team_id, expires_at);
CREATE INDEX IF NOT EXISTS idx_team_file_locks_owner_bot ON team_file_locks(owner_bot_id);
