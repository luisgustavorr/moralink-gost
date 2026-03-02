package updater

import (
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

type Release struct {
	TagName string  `json:"tag_name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

const apiBase = "https://api.github.com"

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"url"`
	Size               int64  `json:"size"`
}

func serviceExecutablePath() string {
	if runtime.GOOS == "windows" {
		return `C:\Program Files\MoraLink\moralink-gost.exe`
	}
	return filepath.Join(os.TempDir(), "moralink-gost")
}

func DownloadRelease(tag string) error {
	release, err := GetRelease(tag)
	asset := pickAsset(release.Assets)

	if asset == nil {
		return fmt.Errorf("no compatible asset found for %s/%s in release %s")
	}
	fmt.Println("Downloading %s (%s)...", asset.Name, asset.BrowserDownloadURL)
	targetPath := serviceExecutablePath()
	tmpPath := targetPath + ".tmp"
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("Authorization", "token "+os.Getenv("RELEASE_GH"))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s\n%s", resp.Status)
	}
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	err = os.Chmod(tmpPath, 0755)

	if runtime.GOOS == "windows" {
		backupPath := targetPath + ".old"
		os.Remove(backupPath)
		if err := os.Rename(targetPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup old binary: %w", err)
		}
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("Binary replaced at %s \n", targetPath)

	return err
}
func pickAsset(assets []Asset) *Asset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	for _, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, goos) && strings.Contains(name, goarch) {
			return &a
		}
	}
	return nil
}
func GetRelease(tag string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", apiBase, "luisgustavorr", "moralink-gost", tag)
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "moralink-updater")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("RELEASE_GH"))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach GitHub: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}
	return &release, nil
}

func isNewer(latest, current string) bool {
	return latest > current
}
