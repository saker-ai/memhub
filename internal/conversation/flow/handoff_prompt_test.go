package flow

import (
	"strings"
	"testing"

	"github.com/memohai/memoh/internal/agentteam"
)

func TestBuildHandoffPromptAnnouncesAuthorAndIssueOnMention(t *testing.T) {
	t.Parallel()

	handoff := agentteam.Handoff{
		FromActorType: agentteam.ActorUser,
		ToBotID:       "bot-2",
	}
	comment := agentteam.Comment{
		AuthorType:   agentteam.ActorUser,
		AuthorUserID: "user-1",
		Content:      "Please draft the spec.",
	}
	issue := &agentteam.Issue{Number: 7, Title: "Login flow"}

	got := buildHandoffPrompt(handoff, comment, "Alice", issue)

	for _, want := range []string{
		"**Alice**",
		"@mentioned you",
		`issue #7 "Login flow"`,
		"## Message from Alice",
		"Please draft the spec.",
		"Address them by name when you reply",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in prompt:\n%s", want, got)
		}
	}
}

func TestBuildHandoffPromptForReturnFlagsDelegation(t *testing.T) {
	t.Parallel()

	handoff := agentteam.Handoff{
		FromActorType: agentteam.ActorSystem,
		ToBotID:       "bot-1",
	}
	comment := agentteam.Comment{
		AuthorType:  agentteam.ActorBot,
		AuthorBotID: "bot-2",
		Content:     "Done, see /team/output.md",
	}
	issue := &agentteam.Issue{Number: 7, Title: "Login flow"}

	got := buildHandoffPrompt(handoff, comment, "Worker", issue)

	for _, want := range []string{
		"Your earlier delegation",
		"**Worker** just posted an update",
		`issue #7 "Login flow"`,
		"## Update from Worker",
		"Done, see /team/output.md",
		"exit silently",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in prompt:\n%s", want, got)
		}
	}
}

func TestBuildHandoffPromptFallsBackWhenIssueMissing(t *testing.T) {
	t.Parallel()

	got := buildHandoffPrompt(
		agentteam.Handoff{FromActorType: agentteam.ActorUser},
		agentteam.Comment{AuthorType: agentteam.ActorUser, Content: "hi"},
		"Alice",
		nil,
	)
	if !strings.Contains(got, "the team issue") {
		t.Fatalf("expected fallback header when issue is nil, got:\n%s", got)
	}
}
