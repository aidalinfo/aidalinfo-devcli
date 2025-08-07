package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	REPO_URL        = "https://github.com/aidalinfo/aidalinfo-devcli"
	CURRENT_VERSION = "v0.0.14"
)

type UpdateInfo struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	DownloadURL     string `json:"downloadUrl"`
}

// ErrSudoRequired is returned when sudo privileges are required
var ErrSudoRequired = errors.New("sudo_required")

func GetCurrentVersion() string {
	return CURRENT_VERSION
}

// compareVersions compare deux versions sémantiques
// Retourne: 1 si v1 > v2, -1 si v1 < v2, 0 si v1 == v2
func compareVersions(v1, v2 string) int {
	// Supprimer le préfixe 'v' s'il existe
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Séparer par les points
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Comparer chaque partie numériquement
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		var err error

		if i < len(parts1) {
			p1, err = strconv.Atoi(parts1[i])
			if err != nil {
				p1 = 0 // Si ce n'est pas un nombre, traiter comme 0
			}
		}

		if i < len(parts2) {
			p2, err = strconv.Atoi(parts2[i])
			if err != nil {
				p2 = 0 // Si ce n'est pas un nombre, traiter comme 0
			}
		}

		if p1 > p2 {
			return 1
		} else if p1 < p2 {
			return -1
		}
	}

	return 0 // Les versions sont égales
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
				// Utiliser la comparaison de versions sémantiques
				if latestVersion == "" || compareVersions(version, latestVersion) > 0 {
					latestVersion = version
				}
			}
		}
	}

	if latestVersion == "" {
		return &UpdateInfo{
			CurrentVersion:  CURRENT_VERSION,
			UpdateAvailable: false,
		}, nil
	}

	arch := runtime.GOARCH
	osName := runtime.GOOS
	downloadURL := fmt.Sprintf("%s/releases/download/%s/aidalinfo-cli_%s_%s",
		REPO_URL, latestVersion, osName, arch)

	// Une mise à jour est disponible si la version distante est plus récente que la version actuelle
	updateAvailable := compareVersions(latestVersion, CURRENT_VERSION) > 0

	return &UpdateInfo{
		CurrentVersion:  CURRENT_VERSION,
		LatestVersion:   latestVersion,
		UpdateAvailable: updateAvailable,
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
		// Si on ne peut pas renommer, on a besoin de sudo
		if os.IsPermission(err) {
			return ErrSudoRequired
		}
		return fmt.Errorf("erreur lors de la sauvegarde de l'ancien binaire: %v", err)
	}

	err = os.Rename(tmpFilePath, currentExe)
	if err != nil {
		// Restaurer l'ancien binaire
		os.Rename(backupPath, currentExe)
		// Si on ne peut pas renommer, on a besoin de sudo
		if os.IsPermission(err) {
			return ErrSudoRequired
		}
		return fmt.Errorf("erreur lors du remplacement du binaire: %v", err)
	}

	os.Remove(backupPath)

	return nil
}

// PerformUpdateWithPassword performs the update using sudo with the provided password
func PerformUpdateWithPassword(tmpFilePath string, password string) error {
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du chemin de l'exécutable: %v", err)
	}

	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("erreur lors de la résolution du symlink: %v", err)
	}

	backupPath := currentExe + ".backup"
	
	// Créer un script pour effectuer la mise à jour
	script := fmt.Sprintf(`#!/bin/bash
mv "%s" "%s" || exit 1
mv "%s" "%s" || { mv "%s" "%s"; exit 1; }
rm -f "%s"
`, 
		currentExe, backupPath, 
		tmpFilePath, currentExe,
		backupPath, currentExe,
		backupPath)
	
	// Utiliser sudo avec le mot de passe fourni
	cmd := exec.Command("sudo", "-S", "bash", "-c", script)
	cmd.Stdin = strings.NewReader(password + "\n")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Vérifier si c'est une erreur d'authentification
		if strings.Contains(string(output), "incorrect password") || 
		   strings.Contains(string(output), "Désolé, essayez de nouveau") ||
		   strings.Contains(string(output), "Sorry, try again") {
			return fmt.Errorf("mot de passe incorrect")
		}
		return fmt.Errorf("erreur lors de la mise à jour: %v - %s", err, string(output))
	}

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
