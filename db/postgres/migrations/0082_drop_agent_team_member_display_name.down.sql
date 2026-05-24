-- 0082_drop_agent_team_member_display_name (down)
-- Restore the column for rollback. Data is not recoverable.

ALTER TABLE agent_team_members ADD COLUMN IF NOT EXISTS display_name TEXT NOT NULL DEFAULT '';
