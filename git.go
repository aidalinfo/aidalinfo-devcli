package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
    "regexp"
	"os/exec"
)

// listSubmodule : Récupère les sous-modules récursivement et retourne leurs chemins
func listSubmodule(path string) ([]string, error) {
	var results []string

	// Si le chemin est vide, utilise le répertoire courant
	if path == "" {
		path = "."
	}
	fmt.Println(path)
	// Chemin vers .gitmodules
	gitModulesPath := filepath.Join(path, ".gitmodules")

	// Vérifie si le fichier .gitmodules existe
	if _, err := os.Stat(gitModulesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf(".gitmodules introuvable dans %s", path)
	}

	// Ouvre le fichier .gitmodules
	file, err := os.Open(gitModulesPath)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de l'ouverture de .gitmodules : %w", err)
	}
	defer file.Close()

	// Parcourt chaque ligne du fichier pour extraire les sous-modules
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println(line)
		if strings.Contains(line, "path = ") {
			// Extrait le chemin du sous-module
			submodulePath := strings.TrimSpace(strings.Split(line, "=")[1])
			// Construit le chemin complet
			fullPath := filepath.Join(path, submodulePath)
			results = append(results, fullPath)

			// Vérifie les sous-modules imbriqués récursivement
			subResults, err := listSubmodule(fullPath)
			if err == nil {
				results = append(results, subResults...)
			}
		}
	}

	// Vérifie les erreurs de lecture du fichier
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Erreur lors de la lecture de .gitmodules : %w", err)
	}

	return results, nil
}


func cleanSubmodule(submodules []string) ([]string, error)  {
	re := regexp.MustCompile(`[^/]+$`)

    // Extraire uniquement la dernière partie de chaque chemin
    var submoduleNames []string
    for _, submodule := range submodules {
        // Utilise la regex pour trouver la partie après le dernier "/"
        matches := re.FindStringSubmatch(submodule)
        if len(matches) > 0 {
            submoduleNames = append(submoduleNames, matches[0])
        }
    }
	return submoduleNames, nil
}


// Fonction pour exécuter "git status" dans un sous-module
func gitStatus(submodule string) string {
	cmd := exec.Command("git", "-C", submodule, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Erreur : %s", err.Error())
	}
	return string(output)
}


// Fonction pour obtenir la branche actuelle du sous-module
func getCurrentBranch(path string) string {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Erreur : %s", err.Error())
	}
	return strings.TrimSpace(string(output))
}

// Fonction pour obtenir la liste des branches disponibles
func getBranches(path string) []string {
	// Effectuer un git fetch pour récupérer les branches distantes
	fetchCmd := exec.Command("git", "-C", path, "fetch", "--all", "--prune")
	if err := fetchCmd.Run(); err != nil {
		return []string{fmt.Sprintf("Erreur lors du fetch : %s", err.Error())}
	}
	cmd := exec.Command("git", "-C", path, "branch", "-a", "--list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{fmt.Sprintf("Erreur : %s", err.Error())}
	}
	branches := strings.Split(string(output), "\n")
	for i := range branches {
		branches[i] = strings.TrimSpace(branches[i])
	}
	return branches
}