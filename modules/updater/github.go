package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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

var githubToken string

func Configure(token string) {
	log.Println("Getting token from make", token)
	githubToken = token
}
func serviceExecutablePath() string {
	exe, err := os.Executable()
	if err == nil {
		return exe
	}
	if runtime.GOOS == "windows" {
		return `C:\Program Files\MoraLink\moralink-gost.exe`
	}
	return "/usr/local/bin/moralink-gost"
}
func spawnApplyAndExit(tmpPath, targetPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows use cmd.exe to wait a moment then apply
		script := fmt.Sprintf(
			`timeout /t 3 /nobreak >nul && move /Y "%s" "%s" && sc start moralink-gost`,
			tmpPath, targetPath,
		)
		cmd = exec.Command("cmd.exe", "/C", script)
	} else {
		// On Linux/macOS: wait, swap, restart via systemctl
		script := fmt.Sprintf(
			`sleep 3 && mv -f "%s" "%s" && systemctl restart moralink-gost`,
			tmpPath, targetPath,
		)
		cmd = exec.Command("bash", "-c", script)
	}

	// Detach from this process so it survives after we exit
	_ = exe // suppress unused warning
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	setSysProcAttrDetached(cmd) // platform-specific, see below

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to spawn updater: %w", err)
	}

	log.Printf("✅ Updater process spawned (pid %d). Service will restart shortly.\n", cmd.Process.Pid)

	// Signal the service to stop cleanly — the spawned process will restart it
	go func() {
		time.Sleep(1 * time.Second)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(os.Interrupt)
	}()

	return nil
}
func DownloadRelease(tag string) error {
	release, err := GetRelease(tag)
	if err != nil {
		return err
	}
	asset := pickAsset(release.Assets)
	if asset == nil {
		return fmt.Errorf("no compatible asset found for this OS/arch in release %s", tag)
	}

	log.Printf("Downloading %s...\n", asset.Name)

	targetPath := serviceExecutablePath()
	// Download to a .tmp next to the real binary
	tmpPath := targetPath + ".tmp"

	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("Authorization", "token "+githubToken)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status from GitHub: %s", resp.Status)
	}

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	if _, err = io.Copy(out, resp.Body); err != nil {
		out.Close()
		return err
	}
	out.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	log.Println("✅ Download complete. Spawning updater process...")

	// Launch ourselves with --update-apply so we exit the service first,
	// then the detached process swaps the binary and restarts the service.
	return spawnApplyAndExit(tmpPath, targetPath)
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
	req.Header.Set("Authorization", "Bearer "+githubToken)

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
