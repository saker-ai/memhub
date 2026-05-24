package agentteam

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

// DefaultLockTTL defines the default file-lock TTL when callers do not specify one.
const DefaultLockTTL = 0

// Service exposes domain operations over the persistence store.
type Service struct {
	store      Store
	dispatcher *Dispatcher
	logger     *slog.Logger
}

// NewService builds a new agentteam.Service.
func NewService(log *slog.Logger, store Store) *Service {
	if log == nil {
		log = slog.Default()
	}
	s := &Service{
		store:  store,
		logger: log.With(slog.String("service", "agentteam")),
	}
	s.dispatcher = NewDispatcher(log, s)
	return s
}

// Dispatcher returns the comment dispatcher used to route mentions to handoffs.
func (s *Service) Dispatcher() *Dispatcher {
	if s == nil {
		return nil
	}
	return s.dispatcher
}

// SetDispatchTrigger wires the conversation-layer trigger function that
// actually executes a handoff for a target bot.
func (s *Service) SetDispatchTrigger(fn TriggerFunc) {
	if s == nil || s.dispatcher == nil {
		return
	}
	s.dispatcher.SetTrigger(fn)
}

// Store exposes the underlying persistence store for advanced callers
// (handoff dispatcher, fs lock service). Most consumers should use the
// service methods directly.
func (s *Service) Store() Store {
	if s == nil {
		return nil
	}
	return s.store
}

// ── Teams ───────────────────────────────────────────────────────────────────

// CreateTeam creates a new team owned by ownerUserID.
//
// Validates name, sharedDirName, and bot membership constraints.
func (s *Service) CreateTeam(ctx context.Context, input CreateTeamInput) (Team, error) {
	if s.store == nil {
		return Team{}, errors.New("agentteam: store not configured")
	}
	if strings.TrimSpace(input.OwnerUserID) == "" {
		return Team{}, fmt.Errorf("%w: owner_user_id required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return Team{}, fmt.Errorf("%w: name required", ErrInvalidInput)
	}
	if input.SharedDirName != "" {
		if err := ValidateSharedDirName(input.SharedDirName); err != nil {
			return Team{}, err
		}
	}
	return s.store.CreateTeam(ctx, input)
}

// GetTeam returns a team by id.
func (s *Service) GetTeam(ctx context.Context, id string) (Team, error) {
	if s.store == nil {
		return Team{}, errors.New("agentteam: store not configured")
	}
	return s.store.GetTeam(ctx, id)
}

// GetTeamForOwner returns the team only when it belongs to ownerUserID.
func (s *Service) GetTeamForOwner(ctx context.Context, id, ownerUserID string) (Team, error) {
	if s.store == nil {
		return Team{}, errors.New("agentteam: store not configured")
	}
	return s.store.GetTeamForOwner(ctx, id, ownerUserID)
}

// ListTeamsByOwner returns all active teams for ownerUserID.
func (s *Service) ListTeamsByOwner(ctx context.Context, ownerUserID string) ([]Team, error) {
	if s.store == nil {
		return nil, errors.New("agentteam: store not configured")
	}
	return s.store.ListTeamsByOwner(ctx, ownerUserID)
}

// ListAllTeamsByOwner includes archived teams.
func (s *Service) ListAllTeamsByOwner(ctx context.Context, ownerUserID string) ([]Team, error) {
	if s.store == nil {
		return nil, errors.New("agentteam: store not configured")
	}
	return s.store.ListAllTeamsByOwner(ctx, ownerUserID)
}

// ListTeamsForBot returns the teams a bot belongs to.
func (s *Service) ListTeamsForBot(ctx context.Context, botID string) ([]Team, error) {
	if s.store == nil {
		return nil, errors.New("agentteam: store not configured")
	}
	return s.store.ListTeamsForBot(ctx, botID)
}

// UpdateTeam patches mutable team fields.
func (s *Service) UpdateTeam(ctx context.Context, id string, input UpdateTeamInput) (Team, error) {
	if s.store == nil {
		return Team{}, errors.New("agentteam: store not configured")
	}
	if input.SharedDirName != nil && *input.SharedDirName != "" {
		if err := ValidateSharedDirName(*input.SharedDirName); err != nil {
			return Team{}, err
		}
	}
	return s.store.UpdateTeam(ctx, id, input)
}

// ArchiveTeam marks the team as archived (soft delete).
func (s *Service) ArchiveTeam(ctx context.Context, id string) (Team, error) {
	if s.store == nil {
		return Team{}, errors.New("agentteam: store not configured")
	}
	return s.store.ArchiveTeam(ctx, id)
}

// DeleteTeam hard-deletes the team and all its rows.
func (s *Service) DeleteTeam(ctx context.Context, id string) error {
	if s.store == nil {
		return errors.New("agentteam: store not configured")
	}
	return s.store.DeleteTeam(ctx, id)
}

// ── Members ─────────────────────────────────────────────────────────────────

// AddMember adds a bot or user to the team.
func (s *Service) AddMember(ctx context.Context, input CreateMemberInput) (Member, error) {
	if s.store == nil {
		return Member{}, errors.New("agentteam: store not configured")
	}
	if strings.TrimSpace(input.TeamID) == "" {
		return Member{}, fmt.Errorf("%w: team_id required", ErrInvalidInput)
	}
	switch input.MemberType {
	case MemberBot:
		if strings.TrimSpace(input.BotID) == "" {
			return Member{}, fmt.Errorf("%w: bot_id required for bot members", ErrInvalidInput)
		}
		input.UserID = ""
	case MemberUser:
		if strings.TrimSpace(input.UserID) == "" {
			return Member{}, fmt.Errorf("%w: user_id required for user members", ErrInvalidInput)
		}
		input.BotID = ""
	default:
		return Member{}, fmt.Errorf("%w: invalid member_type %q", ErrInvalidInput, input.MemberType)
	}
	return s.store.AddMember(ctx, input)
}

// ListMembers returns all members of a team.
func (s *Service) ListMembers(ctx context.Context, teamID string) ([]Member, error) {
	if s.store == nil {
		return nil, errors.New("agentteam: store not configured")
	}
	return s.store.ListMembers(ctx, teamID)
}

// UpdateMember patches mutable member fields.
func (s *Service) UpdateMember(ctx context.Context, id string, input UpdateMemberInput) (Member, error) {
	if s.store == nil {
		return Member{}, errors.New("agentteam: store not configured")
	}
	return s.store.UpdateMember(ctx, id, input)
}

// RemoveMember deletes a member by id.
func (s *Service) RemoveMember(ctx context.Context, id string) error {
	if s.store == nil {
		return errors.New("agentteam: store not configured")
	}
	return s.store.DeleteMember(ctx, id)
}

// IsBotInTeam returns true when botID belongs to teamID.
func (s *Service) IsBotInTeam(ctx context.Context, teamID, botID string) (bool, error) {
	if s.store == nil {
		return false, errors.New("agentteam: store not configured")
	}
	_, err := s.store.GetMemberByBot(ctx, teamID, botID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ── Issues ──────────────────────────────────────────────────────────────────

// CreateIssue creates a new issue scoped to a team.
func (s *Service) CreateIssue(ctx context.Context, input CreateIssueInput) (Issue, error) {
	if s.store == nil {
		return Issue{}, errors.New("agentteam: store not configured")
	}
	if strings.TrimSpace(input.TeamID) == "" {
		return Issue{}, fmt.Errorf("%w: team_id required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Title) == "" {
		return Issue{}, fmt.Errorf("%w: title required", ErrInvalidInput)
	}
	if input.Status == "" {
		input.Status = StatusTodo
	}
	if input.CreatedByType == "" {
		input.CreatedByType = ActorUser
	}
	return s.store.CreateIssue(ctx, input)
}

// GetIssue returns an issue by id.
func (s *Service) GetIssue(ctx context.Context, id string) (Issue, error) {
	if s.store == nil {
		return Issue{}, errors.New("agentteam: store not configured")
	}
	return s.store.GetIssue(ctx, id)
}

// ListIssuesByTeam returns all issues for a team.
func (s *Service) ListIssuesByTeam(ctx context.Context, teamID string) ([]Issue, error) {
	if s.store == nil {
		return nil, errors.New("agentteam: store not configured")
	}
	return s.store.ListIssuesByTeam(ctx, teamID)
}

// ListOpenIssuesByTeam returns non-terminal issues for a team.
func (s *Service) ListOpenIssuesByTeam(ctx context.Context, teamID string) ([]Issue, error) {
	if s.store == nil {
		return nil, errors.New("agentteam: store not configured")
	}
	return s.store.ListOpenIssuesByTeam(ctx, teamID)
}

// ListIssuesForOwner returns issues across all teams owned by ownerUserID.
func (s *Service) ListIssuesForOwner(ctx context.Context, ownerUserID string) ([]Issue, error) {
	if s.store == nil {
		return nil, errors.New("agentteam: store not configured")
	}
	return s.store.ListIssuesForOwner(ctx, ownerUserID)
}

// UpdateIssue patches mutable issue fields.
func (s *Service) UpdateIssue(ctx context.Context, id string, input UpdateIssueInput) (Issue, error) {
	if s.store == nil {
		return Issue{}, errors.New("agentteam: store not configured")
	}
	return s.store.UpdateIssue(ctx, id, input)
}

// AssignIssue updates the issue's assignee (bot/user/none).
func (s *Service) AssignIssue(ctx context.Context, id string, input AssignIssueInput) (Issue, error) {
	if s.store == nil {
		return Issue{}, errors.New("agentteam: store not configured")
	}
	return s.store.SetIssueAssignee(ctx, id, input)
}

// DeleteIssue removes an issue (cascade comments/handoffs).
func (s *Service) DeleteIssue(ctx context.Context, id string) error {
	if s.store == nil {
		return errors.New("agentteam: store not configured")
	}
	return s.store.DeleteIssue(ctx, id)
}

// ── Comments ────────────────────────────────────────────────────────────────

// PostComment creates a comment on an issue.
func (s *Service) PostComment(ctx context.Context, input CreateCommentInput) (Comment, error) {
	if s.store == nil {
		return Comment{}, errors.New("agentteam: store not configured")
	}
	if strings.TrimSpace(input.IssueID) == "" {
		return Comment{}, fmt.Errorf("%w: issue_id required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.TeamID) == "" {
		issue, err := s.store.GetIssue(ctx, input.IssueID)
		if err != nil {
			return Comment{}, err
		}
		input.TeamID = issue.TeamID
	}
	if input.AuthorType == "" {
		input.AuthorType = ActorUser
	}
	cmt, err := s.store.CreateComment(ctx, input)
	if err != nil {
		return Comment{}, err
	}
	if err := s.store.TouchIssueAfterComment(ctx, cmt.IssueID); err != nil && s.logger != nil {
		s.logger.Warn("touch issue after comment failed", slog.String("issue_id", cmt.IssueID), slog.Any("error", err))
	}
	if s.dispatcher != nil {
		if err := s.dispatcher.HandleComment(ctx, cmt); err != nil && s.logger != nil {
			s.logger.Warn("dispatcher handle comment failed", slog.String("comment_id", cmt.ID), slog.Any("error", err))
		}
	}
	return cmt, nil
}

// ListComments returns the comment thread for an issue.
func (s *Service) ListComments(ctx context.Context, issueID string) ([]Comment, error) {
	if s.store == nil {
		return nil, errors.New("agentteam: store not configured")
	}
	return s.store.ListComments(ctx, issueID)
}

// ── Mentions ───────────────────────────────────────────────────────────────

// Mention represents a parsed @-mention from comment content. The platform
// uses a single source of truth for naming: bots are referenced by their
// own `bots.display_name` and humans by their `users.display_name`. The
// dispatcher resolves these names against the issue's team roster — there
// is intentionally no separate "mention id" namespace.
type Mention struct {
	// Name is the verbatim text after the @, e.g. `FrontendBot` for a
	// bare token mention or `Frontend Bot` for a quoted mention.
	Name string
}

// mentionRe matches one of two forms of @-mention in a comment body:
//
//  1. Bare token   : `@FrontendBot`         (word chars + - . _)
//  2. Quoted form  : `@"Frontend Bot"`      (any non-quote run, for names with spaces)
//
// The `@` may appear anywhere in the text — start, middle, end — as
// long as the character immediately before it is NOT a "word" character
// (letter / digit / underscore). This single rule covers all the cases
// people actually care about:
//
//   - `@Alice please`            ✅ start of line
//   - `cc @Alice`                ✅ whitespace boundary
//   - `(@Alice)` / `[@Alice]`    ✅ punctuation boundary
//   - `bob@example.com`          ✗ letter before `@` (email is safe)
//   - `/cmd@bot`                 ✗ letter before `@` (bot-suffix is safe)
//
// The label after `@` is resolved by the dispatcher against the team's
// roster (bots.display_name + users.display_name) at delivery time, so
// renames cascade automatically.
var mentionRe = regexp.MustCompile(`(?:^|\W)@(?:"([^"]+)"|([A-Za-z0-9_][A-Za-z0-9_.\-]*))`)

// ParseMentions extracts deduplicated @-mention labels from a comment body.
// Names are returned verbatim (case-preserving) so callers can present them
// back to the user; matching to team members is case-insensitive.
func ParseMentions(content string) []Mention {
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool, len(matches))
	out := make([]Mention, 0, len(matches))
	for _, m := range matches {
		var name string
		if len(m) >= 2 && m[1] != "" {
			name = m[1]
		} else if len(m) >= 3 && m[2] != "" {
			name = m[2]
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, Mention{Name: name})
	}
	return out
}

// MatchMember resolves a mention label against the team's member roster.
// Returns the matching Member and true when there is exactly one
// case-insensitive name match. Returns the zero Member and false on
// missing / ambiguous / empty input — callers should treat both cases as
// "no routing happens".
func MatchMember(name string, roster []Member) (Member, bool) {
	target := strings.ToLower(strings.TrimSpace(name))
	if target == "" {
		return Member{}, false
	}
	var matched Member
	count := 0
	for _, m := range roster {
		if strings.ToLower(strings.TrimSpace(m.DisplayName)) == target {
			matched = m
			count++
		}
	}
	if count == 1 {
		return matched, true
	}
	return Member{}, false
}

// ── Shared Dir Validation ──────────────────────────────────────────────────

var sharedDirNameRe = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]{0,63}$`)

// ValidateSharedDirName ensures shared_dir_name is safe for filesystem use.
// Rejects path traversal, absolute paths, system paths, and reserved names.
func ValidateSharedDirName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("%w: shared_dir_name cannot be empty", ErrInvalidInput)
	}
	if strings.ContainsAny(trimmed, "/\\\x00") {
		return fmt.Errorf("%w: shared_dir_name cannot contain path separators", ErrInvalidInput)
	}
	if strings.Contains(trimmed, "..") {
		return fmt.Errorf("%w: shared_dir_name cannot contain '..'", ErrInvalidInput)
	}
	if !sharedDirNameRe.MatchString(trimmed) {
		return fmt.Errorf("%w: shared_dir_name has invalid characters", ErrInvalidInput)
	}
	switch strings.ToLower(trimmed) {
	case ".", "..", "data", "team", "tmp", "etc", "var", "proc", "sys", "dev", "root", "home":
		return fmt.Errorf("%w: shared_dir_name %q is reserved", ErrInvalidInput, trimmed)
	}
	return nil
}
