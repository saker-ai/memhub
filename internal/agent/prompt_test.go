package agent

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateSystemPromptIncludesPlatformIdentitiesInChat(t *testing.T) {
	t.Parallel()

	prompt := GenerateSystemPrompt(SystemPromptParams{
		SessionType:               "chat",
		Now:                       time.Unix(1, 0).UTC(),
		Timezone:                  "UTC",
		PlatformIdentitiesSection: "## Platform Identities\n\n<identity channel=\"telegram\" username=\"@memoh\"/>",
	})

	if !strings.Contains(prompt, "## Platform Identities") {
		t.Fatalf("expected platform identities heading in prompt")
	}
	if !strings.Contains(prompt, `<identity channel="telegram" username="@memoh"/>`) {
		t.Fatalf("expected platform identity XML in prompt")
	}
}

func TestGenerateSystemPromptIncludesDisplayToolsWhenEnabled(t *testing.T) {
	t.Parallel()

	prompt := GenerateSystemPrompt(SystemPromptParams{
		SessionType:    "chat",
		Now:            time.Unix(1, 0).UTC(),
		Timezone:       "UTC",
		DisplayEnabled: true,
	})

	if !strings.Contains(prompt, "## Workspace browser & desktop") {
		t.Fatalf("expected display tools section in prompt")
	}
	if !strings.Contains(prompt, "browser_observe") {
		t.Fatalf("expected browser tool mention in prompt")
	}
}

func TestGenerateSystemPromptOmitsDisplayToolsWhenDisabled(t *testing.T) {
	t.Parallel()

	prompt := GenerateSystemPrompt(SystemPromptParams{
		SessionType: "chat",
		Now:         time.Unix(1, 0).UTC(),
		Timezone:    "UTC",
	})

	if strings.Contains(prompt, "## Workspace browser & desktop") {
		t.Fatalf("expected display tools section to be omitted")
	}
}

func TestGenerateSystemPromptIncludesPlatformIdentitiesInDiscuss(t *testing.T) {
	t.Parallel()

	prompt := GenerateSystemPrompt(SystemPromptParams{
		SessionType:               "discuss",
		Now:                       time.Unix(1, 0).UTC(),
		Timezone:                  "UTC",
		PlatformIdentitiesSection: "## Platform Identities\n\n<identity channel=\"discord\" username=\"@memoh\"/>",
	})

	if !strings.Contains(prompt, "## Platform Identities") {
		t.Fatalf("expected platform identities heading in discuss prompt")
	}
	if !strings.Contains(prompt, `<identity channel="discord" username="@memoh"/>`) {
		t.Fatalf("expected platform identity XML in discuss prompt")
	}
}
