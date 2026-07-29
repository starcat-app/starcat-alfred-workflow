// starcat-alfred 是 Alfred Workflow 的薄适配器。
//
// 它只调用 `starcat search`、转换 Script Filter JSON、缓存公开头像；
// 不读取 Starcat 数据库、Keychain、GitHub Token 或 MCP 凭据。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/starcat-app/starcat-alfred-workflow/internal/alfredjson"
	"github.com/starcat-app/starcat-alfred-workflow/internal/avatarcache"
	"github.com/starcat-app/starcat-alfred-workflow/internal/starcatcli"
)

const maxManifestBytes = 256 << 10

type hydrateManifest struct {
	CacheRoot string   `json:"cache_root"`
	URLs      []string `json:"urls"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		// Alfred 需要 JSON 才能稳定展示错误；只有开发者误用子命令时才写 stderr。
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("Usage: starcat-alfred <search|hydrate|generate-assets>")
	}
	switch args[0] {
	case "search":
		if len(args) != 2 {
			return writeOutput(alfredjson.ErrorOutput(
				"请输入仓库关键词",
				"例如：starcat local RAG",
				ensureFallback(defaultCacheRoot()),
			))
		}
		return runSearch(args[1])
	case "hydrate":
		if len(args) != 2 {
			return errors.New("Usage: starcat-alfred hydrate <manifest>")
		}
		return runHydrate(args[1])
	case "generate-assets":
		if len(args) != 2 {
			return errors.New("Usage: starcat-alfred generate-assets <output.png>")
		}
		return avatarcache.EnsureFallback(args[1])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runSearch(query string) error {
	cacheRoot := defaultCacheRoot()
	cache, err := avatarcache.New(cacheRoot)
	if err != nil {
		return writeOutput(alfredjson.ErrorOutput(
			"无法初始化 Alfred 图标缓存",
			"请检查 Workflow Cache 目录权限",
			ensureFallback(cacheRoot),
		))
	}
	if strings.TrimSpace(query) == "" {
		return writeOutput(alfredjson.ErrorOutput(
			"请输入仓库关键词",
			"例如：starcat local RAG",
			cache.FallbackPath(),
		))
	}

	cliPath, err := starcatcli.Resolve(os.Getenv("starcat_cli_path"))
	if err != nil {
		return writeOutput(alfredjson.ErrorOutput(
			"未找到 Starcat CLI",
			"请安装 CLI，或在 Workflow 配置中选择绝对路径",
			cache.FallbackPath(),
		))
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	result, err := (starcatcli.Client{Path: cliPath}).Search(
		timeoutContext,
		query,
		"all",
		resultLimit(),
	)
	if err != nil {
		title, subtitle := userFacingError(err)
		return writeOutput(alfredjson.ErrorOutput(title, subtitle, cache.FallbackPath()))
	}

	output, pending := alfredjson.Render(result, cache, cache.FallbackPath())
	refreshCount := integerEnvironment("avatar_refresh_count", 0)
	if len(pending) > 0 && refreshCount < 3 {
		if startHydrator(cacheRoot, pending) == nil {
			interval := 0.4
			output.Rerun = &interval
			output.Variables = map[string]string{
				"avatar_refresh_count": strconv.Itoa(refreshCount + 1),
			}
		}
	}
	return writeOutput(output)
}

func runHydrate(manifestPath string) error {
	cleanManifestPath := filepath.Clean(manifestPath)
	expectedRoot := filepath.Clean(defaultCacheRoot())
	if filepath.Dir(cleanManifestPath) != expectedRoot ||
		!strings.HasPrefix(filepath.Base(cleanManifestPath), ".hydrate-") {
		return errors.New("hydrate manifest must be inside the Alfred workflow cache")
	}
	file, err := os.Open(cleanManifestPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var manifest hydrateManifest
	if err := json.NewDecoder(io.LimitReader(file, maxManifestBytes)).Decode(&manifest); err != nil {
		return err
	}
	if filepath.Clean(manifest.CacheRoot) != expectedRoot {
		return errors.New("hydrate manifest cache root does not match the workflow cache")
	}
	defer os.Remove(cleanManifestPath)
	cache, err := avatarcache.New(manifest.CacheRoot)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cache.Hydrate(ctx, manifest.URLs)
	return nil
}

func startHydrator(cacheRoot string, urls []string) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	manifestFile, err := os.CreateTemp(cacheRoot, ".hydrate-*.json")
	if err != nil {
		return err
	}
	manifestPath := manifestFile.Name()
	encoderError := json.NewEncoder(manifestFile).Encode(hydrateManifest{
		CacheRoot: cacheRoot,
		URLs:      urls,
	})
	closeError := manifestFile.Close()
	if encoderError != nil || closeError != nil {
		_ = os.Remove(manifestPath)
		if encoderError != nil {
			return encoderError
		}
		return closeError
	}

	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		_ = os.Remove(manifestPath)
		return err
	}
	defer null.Close()
	command := exec.Command(executable, "hydrate", manifestPath)
	command.Stdin = null
	command.Stdout = null
	command.Stderr = null
	if err := command.Start(); err != nil {
		_ = os.Remove(manifestPath)
		return err
	}
	return command.Process.Release()
}

func defaultCacheRoot() string {
	if value := strings.TrimSpace(os.Getenv("alfred_workflow_cache")); value != "" {
		return filepath.Clean(value)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "com.starcat.alfred")
	}
	return filepath.Join(cache, "com.starcat.alfred")
}

func ensureFallback(cacheRoot string) string {
	path := filepath.Join(cacheRoot, "repo-fallback.png")
	_ = avatarcache.EnsureFallback(path)
	return path
}

func resultLimit() int {
	value := integerEnvironment("result_limit", 30)
	if value < 5 {
		return 5
	}
	if value > 50 {
		return 50
	}
	return value
}

func integerEnvironment(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil {
		return fallback
	}
	return value
}

func userFacingError(err error) (string, string) {
	message := strings.ToLower(err.Error())
	switch stableErrorCode(err.Error()) {
	case "CLI_NOT_PAIRED":
		return "Starcat CLI 尚未配对", "在 Starcat 的 MCP 设置中复制配对命令"
	case "REQUIRES_PRO":
		return "Alfred 集成需要 Starcat Pro", "打开 Starcat 查看 Pro 方案"
	case "MCP_DISABLED":
		return "Starcat MCP Service 未开启", "请在 Starcat 设置中开启 MCP Service"
	case "UPGRADE_REQUIRED":
		return "请升级 Starcat 和 CLI", "当前版本不支持全局仓库搜索"
	case "SEARCH_TIMEOUT":
		return "搜索超时", "请检查 Starcat MCP Service 和网络连接"
	case "SEARCH_FAILED":
		return "搜索失败", "请运行 starcat doctor 检查连接"
	}
	// 兼容尚未升级到稳定错误码的旧 CLI；新版本只走上面的 code 分支。
	switch {
	case strings.Contains(message, "not paired"), strings.Contains(message, "pairing profile"):
		return "Starcat CLI 尚未配对", "在 Starcat 的 MCP 设置中复制配对命令"
	case strings.Contains(message, "pro entitlement"), strings.Contains(message, "requires starcat pro"):
		return "Alfred 集成需要 Starcat Pro", "打开 Starcat 查看 Pro 方案"
	case strings.Contains(message, "connection refused"), strings.Contains(message, "mcp service"):
		return "Starcat MCP Service 未开启", "请在 Starcat 设置中开启 MCP Service"
	case strings.Contains(message, "schema version"), strings.Contains(message, "unknown tool"):
		return "请升级 Starcat 和 CLI", "当前版本不支持全局仓库搜索"
	case errors.Is(err, context.DeadlineExceeded), strings.Contains(message, "deadline exceeded"):
		return "搜索超时", "请检查 Starcat MCP Service 和网络连接"
	default:
		return "搜索失败", "请运行 starcat doctor 检查连接"
	}
}

func stableErrorCode(message string) string {
	const marker = "STARCAT_ERROR "
	index := strings.Index(message, marker)
	if index < 0 {
		return ""
	}
	remainder := message[index+len(marker):]
	if separator := strings.IndexByte(remainder, ':'); separator >= 0 {
		return strings.TrimSpace(remainder[:separator])
	}
	return ""
}

func writeOutput(output alfredjson.Output) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(output)
}
