// Package avatarcache 管理 Alfred 只能读取本地图标这一平台约束。
package avatarcache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	maxImageBytes       = 2 << 20
	maxImageEdge        = 1024
	maxFiles            = 500
	maxTotalBytes       = 50 << 20
	cleanupInterval     = 24 * time.Hour
	staleCleanupLockAge = 5 * time.Minute
)

type Cache struct {
	root             string
	fallbackPath     string
	cleanupStatePath string
	cleanupLockPath  string
	client           *http.Client
}

func New(root string) (*Cache, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("avatar cache root is empty")
	}
	avatarRoot := filepath.Join(root, "avatars", "v1")
	if err := os.MkdirAll(avatarRoot, 0o700); err != nil {
		return nil, fmt.Errorf("create avatar cache: %w", err)
	}
	cache := &Cache{
		root:             avatarRoot,
		fallbackPath:     filepath.Join(root, "repo-fallback.png"),
		cleanupStatePath: filepath.Join(avatarRoot, "cleanup-state.json"),
		cleanupLockPath:  filepath.Join(avatarRoot, ".cleanup.lock"),
	}
	cache.client = &http.Client{
		Timeout: 2 * time.Second,
		CheckRedirect: func(request *http.Request, _ []*http.Request) error {
			if !allowedURL(request.URL) {
				return errors.New("avatar redirect target is not allowed")
			}
			return nil
		},
	}
	if err := EnsureFallback(cache.fallbackPath); err != nil {
		return nil, err
	}
	return cache, nil
}

func (c *Cache) FallbackPath() string { return c.fallbackPath }

// Resolve 从不阻塞下载。冷缓存先返回 fallback，并把 URL 标为 pending。
func (c *Cache) Resolve(rawURL string) (string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil || !allowedURL(parsed) {
		return c.fallbackPath, false
	}
	path := c.pathFor(rawURL)
	info, err := os.Stat(path)
	if err == nil && info.Mode().IsRegular() {
		// 头像过期时仍先展示旧文件，后台刷新不让搜索列表退回 fallback。
		return path, time.Since(info.ModTime()) > 7*24*time.Hour
	}
	return c.fallbackPath, true
}

func (c *Cache) Hydrate(ctx context.Context, rawURLs []string) {
	workers := 4
	if len(rawURLs) < workers {
		workers = len(rawURLs)
	}
	jobs := make(chan string)
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for rawURL := range jobs {
				_ = c.download(ctx, rawURL)
			}
		}()
	}
	for _, rawURL := range unique(rawURLs) {
		select {
		case <-ctx.Done():
			close(jobs)
			group.Wait()
			return
		case jobs <- rawURL:
		}
	}
	close(jobs)
	group.Wait()
	c.cleanupIfDue(time.Now())
}

func (c *Cache) pathFor(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return filepath.Join(c.root, hex.EncodeToString(sum[:])+".png")
}

func (c *Cache) download(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || !allowedURL(parsed) {
		return errors.New("avatar URL is not allowed")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "image/png,image/jpeg")
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("avatar HTTP status %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxImageBytes+1))
	if err != nil {
		return err
	}
	if len(data) > maxImageBytes {
		return errors.New("avatar exceeds size limit")
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return errors.New("avatar is not a decodable PNG or JPEG")
	}
	bounds := decoded.Bounds()
	if bounds.Dx() > maxImageEdge || bounds.Dy() > maxImageEdge {
		return errors.New("avatar dimensions exceed limit")
	}

	target := c.pathFor(rawURL)
	temporary, err := os.CreateTemp(c.root, ".avatar-*.png")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := png.Encode(temporary, decoded); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, target)
}

func allowedURL(value *url.URL) bool {
	if value == nil || value.Scheme != "https" || value.User != nil {
		return false
	}
	switch strings.ToLower(value.Hostname()) {
	case "github.com":
		return strings.HasSuffix(strings.ToLower(value.Path), ".png")
	case "avatars.githubusercontent.com":
		return true
	default:
		return false
	}
}

// cleanupIfDue uses a tiny cross-process claim file because Alfred reruns start separate
// helper processes. The state timestamp is committed before scanning: a crash may defer
// cleanup for one day, but can never make every keystroke rescan the cache directory.
func (c *Cache) cleanupIfDue(now time.Time) bool {
	if !c.cleanupIsDue(now) || !c.acquireCleanupLock(now) {
		return false
	}
	defer os.Remove(c.cleanupLockPath)
	// Another process may have completed cleanup while this process was waiting for a
	// stale lock recovery. Recheck after claiming.
	if !c.cleanupIsDue(now) {
		return false
	}
	if err := c.writeCleanupState(now); err != nil {
		return false
	}
	c.cleanup()
	return true
}

func (c *Cache) cleanupIsDue(now time.Time) bool {
	info, err := os.Stat(c.cleanupStatePath)
	if err != nil {
		return os.IsNotExist(err)
	}
	age := now.Sub(info.ModTime())
	return age >= cleanupInterval
}

func (c *Cache) acquireCleanupLock(now time.Time) bool {
	for attempt := 0; attempt < 2; attempt++ {
		lock, err := os.OpenFile(c.cleanupLockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_ = lock.Close()
			return true
		}
		info, statErr := os.Stat(c.cleanupLockPath)
		if statErr != nil || now.Sub(info.ModTime()) < staleCleanupLockAge {
			return false
		}
		// A helper killed during cleanup can leave the claim behind. Only remove locks
		// old enough that no normal 10-second hydrate can still own them.
		if removeErr := os.Remove(c.cleanupLockPath); removeErr != nil {
			return false
		}
	}
	return false
}

func (c *Cache) writeCleanupState(now time.Time) error {
	temporary, err := os.CreateTemp(c.root, ".cleanup-state-*.json")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := fmt.Fprintf(temporary, "{\"last_cleanup_at\":%q}\n", now.UTC().Format(time.RFC3339)); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Chtimes(temporaryPath, now, now); err != nil {
		return err
	}
	return os.Rename(temporaryPath, c.cleanupStatePath)
}

// cleanup 是低成本机会式淘汰；仅看公开头像文件，不写查询词或仓库元数据。
func (c *Cache) cleanup() {
	entries, err := os.ReadDir(c.root)
	if err != nil {
		return
	}
	type cachedFile struct {
		path    string
		size    int64
		modTime time.Time
	}
	files := make([]cachedFile, 0, len(entries))
	var total int64
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".png") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		file := cachedFile{
			path:    filepath.Join(c.root, entry.Name()),
			size:    info.Size(),
			modTime: info.ModTime(),
		}
		files = append(files, file)
		total += file.size
	}
	sort.Slice(files, func(i, j int) bool { return files[i].modTime.Before(files[j].modTime) })
	for len(files) > maxFiles || total > maxTotalBytes {
		oldest := files[0]
		_ = os.Remove(oldest.path)
		total -= oldest.size
		files = files[1:]
	}
}

// EnsureFallback 生成 256x256 的高对比度 Starcat fallback，避免在源码中维护二进制文件。
func EnsureFallback(path string) error {
	if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 256, 256))
	centerX, centerY := 128, 128
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			dx, dy := x-centerX, y-centerY
			switch {
			case dx*dx+dy*dy <= 108*108:
				canvas.Set(x, y, color.RGBA{R: 34, G: 38, B: 48, A: 255})
			default:
				canvas.Set(x, y, color.RGBA{A: 0})
			}
		}
	}
	// 用粗实心 S 形作为无字体依赖标识，在浅色和深色主题下都有足够对比度。
	gold := color.RGBA{R: 247, G: 190, B: 58, A: 255}
	drawRect(canvas, 76, 66, 181, 91, gold)
	drawRect(canvas, 66, 84, 96, 130, gold)
	drawRect(canvas, 76, 116, 181, 141, gold)
	drawRect(canvas, 161, 132, 191, 178, gold)
	drawRect(canvas, 76, 166, 181, 191, gold)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if err := png.Encode(file, canvas); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func drawRect(canvas *image.RGBA, minX, minY, maxX, maxY int, fill color.RGBA) {
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			canvas.Set(x, y, fill)
		}
	}
}

func unique(values []string) []string {
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
