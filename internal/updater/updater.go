package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultRepo      = "GreyRaphael/online_mermaid"
	defaultUserAgent = "online-mermaid-updater"
)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type ReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

type CheckResult struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	ReleaseTitle   string `json:"release_title"`
	ReleaseNotes   string `json:"release_notes"`
	PublishedAt    string `json:"published_at"`
	ReleaseURL     string `json:"release_url"`
	AssetURL       string `json:"asset_url,omitempty"`
	AssetName      string `json:"asset_name,omitempty"`
	CanAutoUpdate  bool   `json:"can_auto_update"`
	UpdateMessage  string `json:"update_message,omitempty"`
	ManualCommand  string `json:"manual_command,omitempty"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
}

var (
	cacheMu       sync.Mutex
	cachedResult  *CheckResult
	cachedAt      time.Time
	cacheDuration = 60 * time.Second
)

// CompareVersions compares two semver strings (e.g. "v1.0.3" vs "v1.0.4").
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
// "dev" is considered older than any release version.
func CompareVersions(v1, v2 string) int {
	v1Clean := strings.TrimSpace(strings.TrimPrefix(v1, "v"))
	v2Clean := strings.TrimSpace(strings.TrimPrefix(v2, "v"))

	if v1Clean == v2Clean {
		return 0
	}
	if v1Clean == "dev" || v1Clean == "" {
		return -1
	}
	if v2Clean == "dev" || v2Clean == "" {
		return 1
	}

	p1 := parseSemverParts(v1Clean)
	p2 := parseSemverParts(v2Clean)

	for i := 0; i < 3; i++ {
		if p1[i] < p2[i] {
			return -1
		}
		if p1[i] > p2[i] {
			return 1
		}
	}
	return 0
}

func parseSemverParts(s string) [3]int {
	var parts [3]int
	// Strip anything after '-' or '+' (e.g. -dirty, +build)
	if idx := strings.IndexAny(s, "-+"); idx != -1 {
		s = s[:idx]
	}
	segments := strings.Split(s, ".")
	for i := 0; i < len(segments) && i < 3; i++ {
		num, _ := strconv.Atoi(segments[i])
		parts[i] = num
	}
	return parts
}

// FindAsset locates the archive asset and optional sha256 asset for given OS and Arch.
func FindAsset(assets []Asset, targetOS, targetArch, tagName string) (*Asset, *Asset) {
	var archiveAsset *Asset
	var sha256Asset *Asset

	ext := ".tar.gz"
	if targetOS == "windows" {
		ext = ".zip"
	}

	expectedPrefix := fmt.Sprintf("online-mermaid-%s-%s-%s", tagName, targetOS, targetArch)

	for i := range assets {
		name := assets[i].Name
		if strings.HasPrefix(name, expectedPrefix) || (strings.Contains(name, targetOS) && strings.Contains(name, targetArch)) {
			if strings.HasSuffix(name, ext+".sha256") || strings.HasSuffix(name, ".sha256") {
				sha256Asset = &assets[i]
			} else if strings.HasSuffix(name, ext) {
				archiveAsset = &assets[i]
			}
		}
	}

	return archiveAsset, sha256Asset
}

// CheckUpdate checks GitHub for the latest release.
func CheckUpdate(ctx context.Context, currentVersion, repo, proxy string, forceRefresh bool) (*CheckResult, error) {
	cacheMu.Lock()
	if !forceRefresh && cachedResult != nil && time.Since(cachedAt) < cacheDuration {
		res := *cachedResult
		cacheMu.Unlock()
		return &res, nil
	}
	cacheMu.Unlock()

	if repo == "" {
		repo = DefaultRepo
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	if proxy != "" {
		proxy = strings.TrimRight(proxy, "/")
		if strings.Contains(proxy, "%s") {
			apiURL = fmt.Sprintf(proxy, apiURL)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create update check request: %w", err)
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if token := os.Getenv("ONLINE_MERMAID_GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query GitHub releases failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if resp.StatusCode == http.StatusForbidden && strings.Contains(string(body), "rate limit") {
			return nil, errors.New("GitHub API rate limit exceeded. Please try again later")
		}
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release payload: %w", err)
	}

	hasUpdate := CompareVersions(currentVersion, release.TagName) < 0

	archiveAsset, _ := FindAsset(release.Assets, runtime.GOOS, runtime.GOARCH, release.TagName)

	execPath, execErr := os.Executable()
	var isWritable bool
	var isGoRun bool
	if execErr == nil {
		execPath, _ = filepath.EvalSymlinks(execPath)
		isGoRun = strings.Contains(execPath, "go-build")
		isWritable = isDirWritable(filepath.Dir(execPath))
	}

	canAutoUpdate := hasUpdate && archiveAsset != nil && isWritable && !isGoRun
	var updateMsg string
	if !isWritable {
		updateMsg = fmt.Sprintf("Executable directory %q is not writable. Manual update required.", filepath.Dir(execPath))
	} else if isGoRun {
		updateMsg = "Running in dev mode (go run). Self-update is disabled."
	} else if archiveAsset == nil {
		updateMsg = fmt.Sprintf("No precompiled binary found for %s/%s.", runtime.GOOS, runtime.GOARCH)
	}

	var assetURL, assetName, manualCmd string
	if archiveAsset != nil {
		assetURL = archiveAsset.BrowserDownloadURL
		assetName = archiveAsset.Name
		if runtime.GOOS == "windows" {
			manualCmd = fmt.Sprintf("curl -LO %s", assetURL)
		} else {
			manualCmd = fmt.Sprintf("curl -fsSL %s | tar -xz && ./online-mermaid", assetURL)
		}
	}

	result := &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		HasUpdate:      hasUpdate,
		ReleaseTitle:   release.Name,
		ReleaseNotes:   release.Body,
		PublishedAt:    release.PublishedAt.Format(time.RFC3339),
		ReleaseURL:     release.HTMLURL,
		AssetURL:       assetURL,
		AssetName:      assetName,
		CanAutoUpdate:  canAutoUpdate,
		UpdateMessage:  updateMsg,
		ManualCommand:  manualCmd,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
	}

	cacheMu.Lock()
	cachedResult = result
	cachedAt = time.Now()
	cacheMu.Unlock()

	return result, nil
}

// ApplyUpdate downloads the release asset, verifies SHA256 checksum, and replaces the running binary.
func ApplyUpdate(ctx context.Context, downloadURL, sha256URL, proxy string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate current executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("eval symlink of executable: %w", err)
	}
	execDir := filepath.Dir(execPath)

	if !isDirWritable(execDir) {
		return fmt.Errorf("directory %q is not writable", execDir)
	}

	// Apply proxy to URLs if needed
	resolvedDownloadURL := resolveProxyURL(downloadURL, proxy)
	resolvedSha256URL := resolveProxyURL(sha256URL, proxy)

	client := &http.Client{Timeout: 5 * time.Minute}

	// 1. Download SHA256 checksum if provided.
	// Fail closed: when a checksum URL is given but cannot be fetched or parsed,
	// abort the update instead of silently skipping integrity verification.
	var expectedSHA256 string
	if resolvedSha256URL != "" {
		sum, err := fetchSHA256(ctx, client, resolvedSha256URL)
		if err != nil {
			return fmt.Errorf("fetch SHA256 checksum: %w", err)
		}
		expectedSHA256 = sum
	}

	// 2. Download Archive to temporary file
	tmpArchive, err := os.CreateTemp(execDir, ".online-mermaid-archive-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary download file: %w", err)
	}
	tmpArchiveName := tmpArchive.Name()
	defer os.Remove(tmpArchiveName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolvedDownloadURL, nil)
	if err != nil {
		tmpArchive.Close()
		return fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		tmpArchive.Close()
		return fmt.Errorf("download release asset failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpArchive.Close()
		return fmt.Errorf("download release asset returned status %d", resp.StatusCode)
	}

	hasher := sha256.New()
	writer := io.MultiWriter(tmpArchive, hasher)
	if _, err := io.Copy(writer, resp.Body); err != nil {
		tmpArchive.Close()
		return fmt.Errorf("write downloaded asset: %w", err)
	}
	tmpArchive.Close()

	// 3. Verify SHA256
	if expectedSHA256 != "" {
		actualSHA256 := hex.EncodeToString(hasher.Sum(nil))
		if actualSHA256 != expectedSHA256 {
			return fmt.Errorf("SHA256 checksum mismatch (expected %s, got %s)", expectedSHA256, actualSHA256)
		}
	}

	// 4. Extract executable from archive
	tmpBinary, err := os.CreateTemp(execDir, ".online-mermaid-bin-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary binary file: %w", err)
	}
	tmpBinaryName := tmpBinary.Name()
	defer os.Remove(tmpBinaryName)

	if strings.HasSuffix(downloadURL, ".zip") {
		if err := extractFromZip(tmpArchiveName, tmpBinary); err != nil {
			tmpBinary.Close()
			return fmt.Errorf("extract zip archive: %w", err)
		}
	} else {
		if err := extractFromTarGz(tmpArchiveName, tmpBinary); err != nil {
			tmpBinary.Close()
			return fmt.Errorf("extract tar.gz archive: %w", err)
		}
	}
	tmpBinary.Close()

	// 5. Ensure executable permission
	if err := os.Chmod(tmpBinaryName, 0755); err != nil {
		return fmt.Errorf("set executable permissions: %w", err)
	}

	// 6. Atomically replace the running executable
	if runtime.GOOS == "windows" {
		backupPath := execPath + ".old"
		_ = os.Remove(backupPath)
		if err := os.Rename(execPath, backupPath); err != nil {
			return fmt.Errorf("rename existing binary: %w", err)
		}
		if err := os.Rename(tmpBinaryName, execPath); err != nil {
			_ = os.Rename(backupPath, execPath)
			return fmt.Errorf("replace executable: %w", err)
		}
	} else {
		if err := os.Rename(tmpBinaryName, execPath); err != nil {
			return fmt.Errorf("atomic rename executable: %w", err)
		}
	}

	return nil
}

func extractFromTarGz(archivePath string, out io.Writer) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := filepath.Base(header.Name)
		if name == "online-mermaid" || name == "online-mermaid.exe" {
			if _, err := io.Copy(out, tr); err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("online-mermaid binary not found in tar.gz archive")
}

func extractFromZip(archivePath string, out io.Writer) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, file := range zr.File {
		name := filepath.Base(file.Name)
		if name == "online-mermaid" || name == "online-mermaid.exe" {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			_, err = io.Copy(out, rc)
			return err
		}
	}
	return errors.New("online-mermaid binary not found in zip archive")
}

// fetchSHA256 downloads a "<hex>  <filename>" checksum file and returns the
// lowercase hex digest. The response body is always closed.
func fetchSHA256(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create checksum request: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum endpoint returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return "", fmt.Errorf("read checksum body: %w", err)
	}
	parts := strings.Fields(string(data))
	if len(parts) == 0 || len(parts[0]) != 64 {
		return "", fmt.Errorf("malformed checksum payload %q", strings.TrimSpace(string(data)))
	}
	return strings.ToLower(parts[0]), nil
}

func resolveProxyURL(rawURL, proxy string) string {
	if proxy == "" || rawURL == "" {
		return rawURL
	}
	proxy = strings.TrimRight(proxy, "/")
	if strings.Contains(proxy, "%s") {
		return fmt.Sprintf(proxy, rawURL)
	}
	// Default prefix proxy (e.g. "https://ghproxy.net/https://github.com/...")
	return proxy + "/" + rawURL
}

func isDirWritable(dir string) bool {
	tmp, err := os.CreateTemp(dir, ".perm-check-*")
	if err != nil {
		return false
	}
	name := tmp.Name()
	tmp.Close()
	_ = os.Remove(name)
	return true
}
