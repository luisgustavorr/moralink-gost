package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubOwner = "luisgustavorr"
	githubRepo  = "moralink-gost"

	serviceName = "moralink-gost"
	applyArg    = "--apply-update"
)

type Release struct {
	TagName string  `json:"tag_name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

var githubToken string

func Configure(token string) {
	githubToken = token
}

func serviceExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	return filepath.Abs(exe)
}

// RunApplyMode must be called before normal application initialization.
// handled is true whenever the special updater command was supplied.
func RunApplyMode(args []string) (handled bool, err error) {
	if len(args) == 0 || args[0] != applyArg {
		return false, nil
	}

	if len(args) != 3 {
		return true, fmt.Errorf(
			"usage: %s <temporary-file> <target-file>",
			applyArg,
		)
	}

	tmp, err := filepath.Abs(args[1])
	if err != nil {
		return true, err
	}

	target, err := filepath.Abs(args[2])
	if err != nil {
		return true, err
	}

	if filepath.Clean(tmp) == filepath.Clean(target) {
		return true, fmt.Errorf("temporary and target paths must differ")
	}

	if err := applyUpdate(tmp, target); err != nil {
		return true, err
	}

	return true, nil
}

// RequestUpdate downloads the release and schedules the replacement.
// Call this from the update-event handler, not from application startup.
func RequestUpdate(ctx context.Context, tag string) error {
	if tag == "" {
		return fmt.Errorf("empty release tag")
	}

	release, err := getRelease(ctx, tag)
	if err != nil {
		return err
	}

	asset := pickAsset(release.Assets)
	if asset == nil {
		return fmt.Errorf(
			"no compatible asset for %s/%s in release %s",
			runtime.GOOS, runtime.GOARCH, tag,
		)
	}

	if asset.Size <= 0 {
		return fmt.Errorf("release asset has invalid size: %d", asset.Size)
	}

	target, err := serviceExecutablePath()
	if err != nil {
		return err
	}

	tmp := target + ".download.tmp"
	if err := downloadAsset(ctx, asset, tmp); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	// Do not launch the updater if the file is clearly invalid.
	if err := validateDownloadedBinary(tmp, asset.Size); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := launchApply(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("launch update helper: %w", err)
	}

	return nil
}

func getRelease(ctx context.Context, tag string) (*Release, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/tags/%s",
		githubOwner, githubRepo, tag,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "moralink-gost-updater")

	if githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+githubToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub release API: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	if release.TagName != tag {
		return nil, fmt.Errorf(
			"release tag mismatch: requested %q, received %q",
			tag, release.TagName,
		)
	}

	return &release, nil
}

func pickAsset(assets []Asset) *Asset {
	for i := range assets {
		a := &assets[i]
		name := strings.ToLower(a.Name)

		if strings.Contains(name, strings.ToLower(runtime.GOOS)) &&
			strings.Contains(name, strings.ToLower(runtime.GOARCH)) {
			return a
		}
	}
	return nil
}

func downloadAsset(ctx context.Context, asset *Asset, dst string) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, asset.BrowserDownloadURL, nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "moralink-gost-updater")
	if githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+githubToken)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("asset download: %s", resp.Status)
	}

	f, err := os.OpenFile(
		dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600,
	)
	if err != nil {
		return err
	}

	// Always close the file, and report write/close failures.
	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()

	if copyErr != nil {
		return fmt.Errorf("write downloaded asset: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close downloaded asset: %w", closeErr)
	}

	return nil
}

func validateDownloadedBinary(path string, expectedSize int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("download is not a regular file")
	}
	if info.Size() != expectedSize {
		return fmt.Errorf(
			"download size mismatch: got %d, expected %d",
			info.Size(), expectedSize,
		)
	}
	if info.Size() == 0 {
		return fmt.Errorf("download is empty")
	}

	// Read the whole file to ensure it is readable and calculate its digest.
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	_ = hex.EncodeToString(h.Sum(nil)) // Log/store for comparison if desired.
	return nil
}

// Compile-time reference ensures accidental removal of the import is caught
// if the implementation is refactored to use exec in this shared file.

func combineErrors(primary error, others ...error) error {
	for _, err := range others {
		if err != nil {
			primary = fmt.Errorf("%w; additional failure: %v", primary, err)
		}
	}
	return primary
}

func validatePaths(tmp, target string) error {
	if filepath.Dir(tmp) != filepath.Dir(target) {
		return fmt.Errorf("temporary file must be beside target executable")
	}

	info, err := os.Stat(tmp)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return fmt.Errorf("temporary executable is invalid")
	}

	return nil
}
