package alfredjson

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/starcat-app/starcat-alfred-workflow/internal/starcatcli"
)

var repositorySegmentPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

type IconResolver interface {
	Resolve(rawURL string) (path string, pending bool)
}

// Render 把产品级搜索契约映射为 Alfred 展示模型，不包含任何网络 I/O。
func Render(result starcatcli.Result, icons IconResolver, fallbackPath string) (Output, []string) {
	items := make([]Item, 0, len(result.Items)+len(result.Warnings))
	pendingURLs := make([]string, 0)

	for _, repository := range result.Items {
		openURL, ok := safeOpenURL(repository.OpenURL)
		if !ok {
			continue
		}
		iconPath := fallbackPath
		if repository.IconURL != "" && icons != nil {
			if resolved, pending := icons.Resolve(repository.IconURL); resolved != "" {
				iconPath = resolved
				if pending {
					pendingURLs = append(pendingURLs, repository.IconURL)
				}
			}
		}

		items = append(items, Item{
			UID:          repositoryUID(repository),
			Title:        repository.FullName,
			Subtitle:     subtitle(repository),
			Arg:          openURL,
			Autocomplete: repository.FullName,
			Match:        matchText(repository),
			Valid:        true,
			Icon:         Icon{Path: iconPath},
		})
	}

	if len(items) == 0 {
		items = append(items, Item{
			Title:    "没有找到仓库",
			Subtitle: "已搜索 Starcat 本地仓库和 GitHub",
			Valid:    false,
			Icon:     Icon{Path: fallbackPath},
		})
	}
	for _, warning := range result.Warnings {
		items = append(items, Item{
			Title:    "部分搜索来源暂不可用",
			Subtitle: collapseText(warning, 160),
			Valid:    false,
			Icon:     Icon{Path: fallbackPath},
		})
	}
	return Output{Items: items}, uniqueStrings(pendingURLs)
}

func ErrorOutput(title, subtitle, fallbackPath string) Output {
	return Output{Items: []Item{{
		Title: title, Subtitle: subtitle, Valid: false, Icon: Icon{Path: fallbackPath},
	}}}
}

func repositoryUID(repository starcatcli.Repository) string {
	if repository.RepoID != nil {
		return "repo:" + strconv.FormatInt(*repository.RepoID, 10)
	}
	return "repo-name:" + strings.ToLower(repository.FullName)
}

func subtitle(repository starcatcli.Repository) string {
	parts := []string{"GitHub"}
	if repository.PrimarySource == "local" {
		parts[0] = "Starcat 本地"
	}
	if repository.Language != nil && strings.TrimSpace(*repository.Language) != "" {
		parts = append(parts, strings.TrimSpace(*repository.Language))
	}
	parts = append(parts, "★ "+shortNumber(repository.StarsCount))
	if repository.Description != nil {
		if description := collapseText(*repository.Description, 140); description != "" {
			parts = append(parts, description)
		}
	}
	return strings.Join(parts, " · ")
}

func matchText(repository starcatcli.Repository) string {
	parts := []string{repository.Owner, repository.Name, repository.FullName}
	if repository.Description != nil {
		parts = append(parts, collapseText(*repository.Description, 300))
	}
	if repository.Language != nil {
		parts = append(parts, strings.TrimSpace(*repository.Language))
	}
	return strings.Join(parts, " ")
}

func safeOpenURL(raw string) (string, bool) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil || parsed.Fragment != "" || !validRepositoryPath(parsed.Path) {
		return "", false
	}
	switch parsed.Scheme {
	case "starcat":
		if parsed.Host != "repo" || parsed.Port() != "" {
			return "", false
		}
		query, err := url.ParseQuery(parsed.RawQuery)
		if err != nil || len(query["v"]) != 1 || query.Get("v") != "1" {
			return "", false
		}
		for key := range query {
			if key != "v" && key != "rid" {
				return "", false
			}
		}
		if values, exists := query["rid"]; exists {
			if len(values) != 1 {
				return "", false
			}
			repoID, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil || repoID <= 0 {
				return "", false
			}
		}
		return raw, true
	case "https":
		return raw, strings.EqualFold(parsed.Hostname(), "github.com") &&
			parsed.Port() == "" &&
			parsed.RawQuery == ""
	default:
		return "", false
	}
}

func validRepositoryPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 2 &&
		repositorySegmentPattern.MatchString(parts[0]) &&
		repositorySegmentPattern.MatchString(parts[1])
}

func collapseText(value string, maxRunes int) string {
	value = strings.Join(strings.Fields(value), " ")
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxRunes-1]) + "…"
}

func shortNumber(value int) string {
	absolute := math.Abs(float64(value))
	switch {
	case absolute >= 1_000_000:
		return trimDecimal(float64(value)/1_000_000) + "M"
	case absolute >= 1_000:
		return trimDecimal(float64(value)/1_000) + "k"
	default:
		return strconv.Itoa(value)
	}
}

func trimDecimal(value float64) string {
	if math.Abs(value) >= 100 {
		return fmt.Sprintf("%.0f", value)
	}
	return strings.TrimSuffix(fmt.Sprintf("%.1f", value), ".0")
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
