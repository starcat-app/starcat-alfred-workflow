package alfredjson

import (
	"testing"

	"github.com/starcat-app/starcat-alfred-workflow/internal/starcatcli"
)

func TestRenderShowsSourceAndUsesSafeOpenURL(t *testing.T) {
	repoID := int64(42)
	description := "Local-first\nRAG toolkit"
	language := "Swift"
	result := starcatcli.Result{
		SchemaVersion: 1,
		Items: []starcatcli.Repository{{
			RepoID: &repoID, Owner: "openai", Name: "codex", FullName: "openai/codex",
			Description: &description, Language: &language, StarsCount: 12_400,
			PrimarySource: "local", Sources: []string{"local", "github"},
			IconURL: "https://github.com/openai.png?size=80",
			OpenURL: "starcat://repo/openai/codex?v=1&rid=42",
		}},
	}
	output, pending := Render(result, stubIcons{}, "/fallback.png")

	if len(output.Items) != 1 || !output.Items[0].Valid {
		t.Fatalf("items = %#v", output.Items)
	}
	if output.Items[0].UID != "repo:42" {
		t.Fatalf("uid = %q", output.Items[0].UID)
	}
	if output.Items[0].Subtitle != "Starcat 本地 · Swift · ★ 12.4k · Local-first RAG toolkit" {
		t.Fatalf("subtitle = %q", output.Items[0].Subtitle)
	}
	if output.Items[0].Arg != "starcat://repo/openai/codex?v=1&rid=42" {
		t.Fatalf("arg = %q", output.Items[0].Arg)
	}
	if len(pending) != 1 {
		t.Fatalf("pending = %#v", pending)
	}
}

func TestRenderDropsUnsafeOpenURL(t *testing.T) {
	result := starcatcli.Result{
		SchemaVersion: 1,
		Items: []starcatcli.Repository{{
			Owner: "owner", Name: "repo", FullName: "owner/repo",
			PrimarySource: "github", OpenURL: "file:///tmp/private",
		}},
	}
	output, _ := Render(result, stubIcons{}, "/fallback.png")
	if len(output.Items) != 1 || output.Items[0].Valid || output.Items[0].Title != "没有找到仓库" {
		t.Fatalf("unsafe URL output = %#v", output.Items)
	}
}

func TestSafeOpenURLRejectsNonRepositoryTargets(t *testing.T) {
	invalid := []string{
		"https://github.com/openai/codex/issues",
		"https://user@github.com/openai/codex",
		"https://github.com/openai/codex?tab=readme",
		"starcat://repo/openai/codex?v=2&rid=42",
		"starcat://repo/openai/codex?v=1&rid=0",
		"starcat://repo/openai/codex?v=1&unexpected=true",
		"starcat://repo/openai/codex/extra?v=1",
	}
	for _, value := range invalid {
		if _, ok := safeOpenURL(value); ok {
			t.Fatalf("safeOpenURL(%q) = true, want false", value)
		}
	}
}

func TestSafeOpenURLAcceptsRepositoryTargets(t *testing.T) {
	valid := []string{
		"https://github.com/openai/codex",
		"starcat://repo/openai/codex?v=1",
		"starcat://repo/openai/codex?v=1&rid=42",
	}
	for _, value := range valid {
		if resolved, ok := safeOpenURL(value); !ok || resolved != value {
			t.Fatalf("safeOpenURL(%q) = %q, %v", value, resolved, ok)
		}
	}
}

type stubIcons struct{}

func (stubIcons) Resolve(string) (string, bool) { return "/fallback.png", true }
