package starcatcli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchPassesQueryAsSingleArgument(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "starcat")
	script := `#!/bin/sh
if [ "$1" != "search" ] || [ "$2" != "RAG; echo unsafe" ]; then
  exit 9
fi
printf '%s\n' '{"schema_version":1,"query":"RAG","returned_count":0,"items":[],"providers":{},"warnings":[]}'
`
	if err := os.WriteFile(executable, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := (Client{Path: executable}).Search(
		context.Background(),
		"RAG; echo unsafe",
		"all",
		30,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.SchemaVersion != 1 {
		t.Fatalf("schema version = %d", result.SchemaVersion)
	}
}

func TestResolveRejectsRelativeConfiguredPath(t *testing.T) {
	if _, err := Resolve("./starcat"); err == nil {
		t.Fatal("Resolve() accepted a relative configured path")
	}
}
