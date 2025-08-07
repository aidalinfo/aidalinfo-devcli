package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CleanSubmoduleName nettoie le nom d'un submodule en extrayant uniquement le dernier segment du chemin
func CleanSubmoduleName(submodule string) (string, error) {
	re := regexp.MustCompile(`[^/\\]+$`)
	matches := re.FindStringSubmatch(submodule)
	if len(matches) > 0 {
		return matches[0], nil
	}
	return filepath.Base(submodule), nil
}

// NpmUpdateAction met à jour les dépendances NPM avec le path en paramètre
func NpmUpdateAction(path string) error {
	initialDir, err := os.Getwd()
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur récupération répertoire courant: %v", err))
		return fmt.Errorf("erreur lors de la récupération du répertoire courant: %v", err)
	}
	
	if path != "" && path != "." {
		if err := os.Chdir(path); err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur changement de répertoire vers %s: %v", path, err))
			return fmt.Errorf("erreur lors du changement de répertoire vers %s: %v", path, err)
		}
	}
	defer os.Chdir(initialDir)

	submodules, err := ListSubmodule(".")
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur lors de la liste des submodules: %v", err))
		return fmt.Errorf("erreur lors de la liste des submodules: %v", err)
	}

	LogToFrontend("info", fmt.Sprintf("Mise à jour NPM pour %d submodules", len(submodules)))
	
	for _, submodule := range submodules {
		packageJSONPath := filepath.Join(submodule, "package.json")
		if _, err := os.Stat(packageJSONPath); !os.IsNotExist(err) {
			LogToFrontend("info", fmt.Sprintf("Mise à jour NPM dans %s", submodule))
			if err := execCommand("npm", "-C", submodule, "update"); err != nil {
				LogToFrontend("warning", fmt.Sprintf("Échec mise à jour NPM dans %s: %v", submodule, err))
			}
		}
	}
	
	return nil
}

// GitUpdateAction met à jour les sous-modules avec le path en paramètre
func GitUpdateAction(path string, submodules []string) error {
	initialDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du répertoire courant: %v", err)
	}
	
	if path != "" && path != "." {
		if err := os.Chdir(path); err != nil {
			return fmt.Errorf("erreur lors du changement de répertoire vers %s: %v", path, err)
		}
	}
	defer os.Chdir(initialDir)

	LogToFrontend("info", fmt.Sprintf("Mise à jour git pour %d submodules", len(submodules)))
	
	for _, submodule := range submodules {
		// Nettoyer le chemin du submodule pour éviter les doubles slashes
		cleanPath := strings.TrimPrefix(submodule, path+"/")
		cleanPath = strings.TrimPrefix(cleanPath, "./")
		
		LogToFrontend("info", fmt.Sprintf("Git pull dans %s", cleanPath))
		if err := execCommand("git", "-C", cleanPath, "pull"); err != nil {
			LogToFrontend("warning", fmt.Sprintf("Échec git pull dans %s: %v", cleanPath, err))
		}
	}
	
	return nil
}

// GetCurrentBranch récupère la branche courante avec gestion d'erreur
func GetCurrentBranch(path string) (string, error) {
	output, err := execCommandOutput("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("erreur lors de la récupération de la branche: %v", err)
	}
	return strings.TrimSpace(output), nil
}