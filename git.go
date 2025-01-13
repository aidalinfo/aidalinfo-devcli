package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
    "regexp"
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