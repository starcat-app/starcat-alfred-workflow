// Package starcatcli 只负责定位并调用 Starcat CLI，不读取数据库或任何凭据。
package starcatcli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result 是 starcat.global_search_repos 的稳定 JSON 契约。
type Result struct {
	SchemaVersion int          `json:"schema_version"`
	Query         string       `json:"query"`
	ReturnedCount int          `json:"returned_count"`
	Items         []Repository `json:"items"`
	Providers     Providers    `json:"providers"`
	Warnings      []string     `json:"warnings"`
}

type Repository struct {
	RepoID        *int64   `json:"repo_id"`
	Owner         string   `json:"owner"`
	Name          string   `json:"name"`
	FullName      string   `json:"full_name"`
	Description   *string  `json:"description"`
	Language      *string  `json:"language"`
	StarsCount    int      `json:"stars_count"`
	IsPrivate     bool     `json:"is_private"`
	IsStarred     bool     `json:"is_starred"`
	PrimarySource string   `json:"primary_source"`
	Sources       []string `json:"sources"`
	IconURL       string   `json:"icon_url"`
	OpenURL       string   `json:"open_url"`
	HTMLURL       string   `json:"html_url"`
	UpdatedAt     *string  `json:"updated_at"`
}

type ProviderState struct {
	Status  string  `json:"status"`
	Count   int     `json:"count"`
	Message *string `json:"message"`
}

type Providers struct {
	Local  *ProviderState `json:"local"`
	GitHub *ProviderState `json:"github"`
}

// Client 通过 argv 调用 CLI，查询词永远不会拼入 shell 字符串。
type Client struct {
	Path string
}

func Resolve(explicitPath string) (string, error) {
	candidates := make([]string, 0, 5)
	if explicitPath = strings.TrimSpace(explicitPath); explicitPath != "" {
		if !filepath.IsAbs(explicitPath) {
			return "", errors.New("configured Starcat CLI path must be absolute")
		}
		candidates = append(candidates, explicitPath)
	}
	if path, err := exec.LookPath("starcat"); err == nil {
		candidates = append(candidates, path)
	}
	candidates = append(candidates,
		"/opt/homebrew/bin/starcat",
		"/usr/local/bin/starcat",
	)
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".local", "bin", "starcat"))
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		info, err := os.Stat(candidate)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
			continue
		}
		return candidate, nil
	}
	return "", ErrCLINotFound
}

var ErrCLINotFound = errors.New("Starcat CLI was not found")

func (c Client) Search(ctx context.Context, query, source string, limit int) (Result, error) {
	command := exec.CommandContext(
		ctx,
		c.Path,
		"search",
		query,
		"--source",
		source,
		"--limit",
		fmt.Sprintf("%d", limit),
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return Result{}, fmt.Errorf("starcat search failed: %s", message)
	}

	var result Result
	decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	if err := decoder.Decode(&result); err != nil {
		return Result{}, fmt.Errorf("Starcat CLI returned invalid search JSON: %w", err)
	}
	if result.SchemaVersion != 1 {
		return Result{}, fmt.Errorf("unsupported Starcat search schema version %d", result.SchemaVersion)
	}
	return result, nil
}
