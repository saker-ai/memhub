-- 0082_drop_agent_team_member_display_name
-- Drop the redundant `display_name` column from agent_team_members.
-- The canonical name lives on the underlying bots.display_name / users.display_name
-- and is read at query time via a LEFT JOIN — see db/postgres/queries/agent_teams.sql.

ALTER TABLE agent_team_members DROP COLUMN IF EXISTS display_name;
