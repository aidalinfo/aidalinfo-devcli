package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	REPO_URL     = "https://github.com/aidalinfo/aidalinfo-devcli"
	CURRENT_VERSION = "1.0.0"
)

type UpdateInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	DownloadURL    string `json:"downloadUrl"`
}

func GetCurrentVersion() string {
	return CURRENT_VERSION
}

func CheckForUpdates() (*UpdateInfo, error) {
	cmd := exec.Command("git", "ls-remote", "--tags", REPO_URL)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des tags: %v", err)
	}

	tags := strings.Split(string(output), "\n")
	var latestVersion string
	for _, tag := range tags {
		if strings.Contains(tag, "refs/tags/") && !strings.Contains(tag, "^{}") {
			parts := strings.Split(tag, "refs/tags/")
			if len(parts) == 2 {
				version := strings.TrimSpace(parts[1])
				if latestVersion == "" || version > latestVersion {
					latestVersion = version
				}
			}
		}
	}

	if latestVersion == "" {
		return &UpdateInfo{
			CurrentVersion: CURRENT_VERSION,
			UpdateAvailable: false,
		}, nil
	}

	latestVersionClean := strings.TrimPrefix(latestVersion, "v")
	currentVersionClean := strings.TrimPrefix(CURRENT_VERSION, "v")

	arch := runtime.GOARCH
	osName := runtime.GOOS
	downloadURL := fmt.Sprintf("%s/releases/download/%s/aidalinfo-cli_%s_%s", 
		REPO_URL, latestVersion, osName, arch)

	return &UpdateInfo{
		CurrentVersion:  CURRENT_VERSION,
		LatestVersion:   latestVersion,
		UpdateAvailable: latestVersionClean != currentVersionClean,
		DownloadURL:     downloadURL,
	}, nil
}

func DownloadUpdate(downloadURL string) (string, error) {
	tmpFile, err := os.CreateTemp("", "aidalinfo-cli-update-*")
	if err != nil {
		return "", fmt.Errorf("erreur lors de la création du fichier temporaire: %v", err)
	}
	defer tmpFile.Close()

	resp, err := http.Get(downloadURL)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("erreur lors du téléchargement: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("erreur lors du téléchargement, status: %d", resp.StatusCode)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("erreur lors de l'écriture du fichier: %v", err)
	}

	err = os.Chmod(tmpFile.Name(), 0755)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("erreur lors du chmod: %v", err)
	}

	return tmpFile.Name(), nil
}

func PerformUpdate(tmpFilePath string) error {
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du chemin de l'exécutable: %v", err)
	}
	
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("erreur lors de la résolution du symlink: %v", err)
	}

	backupPath := currentExe + ".backup"
	err = os.Rename(currentExe, backupPath)
	if err != nil {
		cmd := exec.Command("sudo", "mv", currentExe, backupPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("erreur lors de la sauvegarde de l'ancien binaire: %v", err)
		}
	}

	err = os.Rename(tmpFilePath, currentExe)
	if err != nil {
		cmd := exec.Command("sudo", "mv", tmpFilePath, currentExe)
		if err := cmd.Run(); err != nil {
			os.Rename(backupPath, currentExe)
			return fmt.Errorf("erreur lors du remplacement du binaire: %v", err)
		}
	}

	os.Remove(backupPath)
	
	return nil
}

func GetLatestReleaseInfo() (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/aidalinfo/aidalinfo-devcli/releases/latest")
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var releaseInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		return nil, err
	}

	return releaseInfo, nil
}