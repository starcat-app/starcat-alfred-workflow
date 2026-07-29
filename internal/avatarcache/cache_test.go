package avatarcache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveAllowsOnlyGitHubAvatarHosts(t *testing.T) {
	cache, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if path, pending := cache.Resolve("https://github.com/openai.png?size=80"); !pending || path != cache.FallbackPath() {
		t.Fatalf("GitHub avatar = path %q pending %v", path, pending)
	}
	if path, pending := cache.Resolve("https://example.com/avatar.png"); pending || path != cache.FallbackPath() {
		t.Fatalf("untrusted avatar = path %q pending %v", path, pending)
	}
	if path, pending := cache.Resolve("file:///tmp/avatar.png"); pending || path != cache.FallbackPath() {
		t.Fatalf("file avatar = path %q pending %v", path, pending)
	}
}

func TestCleanupRunsAtMostOncePerDay(t *testing.T) {
	cache, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.July, 29, 12, 0, 0, 0, time.UTC)
	if !cache.cleanupIfDue(now) {
		t.Fatal("first cleanup should run")
	}
	if cache.cleanupIfDue(now.Add(23 * time.Hour)) {
		t.Fatal("cleanup should not run again within 24 hours")
	}
	if !cache.cleanupIfDue(now.Add(25 * time.Hour)) {
		t.Fatal("cleanup should run after 24 hours")
	}
}

func TestEnsureFallbackWritesDecodablePNG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "assets", "fallback.png")
	if err := EnsureFallback(path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		t.Fatalf("fallback stat = %#v, %v", info, err)
	}
}
