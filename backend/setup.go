package backend

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SubmoduleAction effectue le checkout des submodules dans le chemin donné
func SubmoduleAction(path string, branches ...string) error {
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

	LogToFrontend("info", fmt.Sprintf("On est dans le répertoire %s", path))

	LogToFrontend("info", "On initialise et update les submodules")
	if err := execCommand("git", "submodule", "init"); err != nil {
		LogToFrontend("error", "Erreur git submodule init")
		return err
	}
	if err := execCommand("git", "submodule", "update"); err != nil {
		LogToFrontend("error", "Erreur git submodule update")
		return err
	}

	defaultBranch, err := GetDefaultBranch()
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur récupération branche par défaut: %v", err))
		return fmt.Errorf("erreur lors de la récupération de la branche par défaut : %v", err)
	}
	branches = append(branches, defaultBranch)

	LogToFrontend("info", fmt.Sprintf("Branches à essayer : %v", branches))

	for _, branch := range branches {
		LogToFrontend("info", fmt.Sprintf("Tentative de checkout de la branche '%s'", branch))
		if err := execCommand("git", "checkout", branch); err == nil {
			LogToFrontend("success", fmt.Sprintf("Branche '%s' checkoutée avec succès", branch))
			break
		}
		LogToFrontend("warn", fmt.Sprintf("Impossible de checkout '%s'", branch))
	}

	LogToFrontend("info", "On pull")
	if err := execCommand("git", "pull"); err != nil {
		LogToFrontend("error", "Erreur git pull")
		return err
	}

	content, err := os.ReadFile(".gitmodules")
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur lecture .gitmodules: %v", err))
		return fmt.Errorf("erreur lors de la lecture de .gitmodules: %v", err)
	}

	var submodules []string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "path = ") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[1])
				submodules = append(submodules, path)
			}
		}
	}

	LogToFrontend("info", fmt.Sprintf("Submodules trouvés : %v", submodules))

	for _, submodule := range submodules {
		LogToFrontend("info", fmt.Sprintf("On entre dans le submodule: %s", submodule))
		absSubmodulePath := filepath.Join(path, submodule)
		LogToFrontend("info", fmt.Sprintf("On va dans le répertoire %s", absSubmodulePath))

		if err := os.Chdir(absSubmodulePath); err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur changement de répertoire: chdir %s: %v", absSubmodulePath, err))
			return fmt.Errorf("erreur lors du changement de répertoire: chdir %s: %v", absSubmodulePath, err)
		}

		for _, branch := range branches {
			LogToFrontend("info", fmt.Sprintf("Tentative de checkout de la branche '%s' pour le submodule", branch))
			if err := execCommand("git", "checkout", branch); err == nil {
				LogToFrontend("success", fmt.Sprintf("Submodule sur branche '%s' checkouté avec succès", branch))
				break
			}
			LogToFrontend("warn", fmt.Sprintf("Submodule : Impossible de checkout '%s'", branch))
		}

		LogToFrontend("info", "On pull (submodule)")
		if err := execCommand("git", "pull"); err != nil {
			LogToFrontend("error", "Erreur git pull (submodule)")
			return err
		}

		if _, err := os.Stat(".gitmodules"); err == nil {
			LogToFrontend("info", "Submodule contient un .gitmodules, récursivité !")
			if err := SubmoduleAction(absSubmodulePath, branches...); err != nil {
				return err
			}
		}

		if err := os.Chdir(initialDir); err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur retour répertoire parent: %v", err))
			return fmt.Errorf("erreur lors du retour au répertoire parent: %v", err)
		}
	}

	return nil
}

// NpmAction lance npm install récursif à partir du path donné si all == true
func NpmAction(path string, all bool) error {
	if !all {
		return nil
	}
	return npmInstallRecursive(path)
}

func npmInstallRecursive(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		LogToFrontend("warn", fmt.Sprintf("Impossible de lire le répertoire %s (permissions?): %v - on continue", path, err))
		return nil // On continue même si on ne peut pas lire le répertoire
	}

	// Si package.json existe dans ce dossier, on fait npm install
	packageJsonPath := filepath.Join(path, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		LogToFrontend("info", fmt.Sprintf("%s : package.json existe, lancement de 'npm install'...", path))
		cmd := exec.Command("npm", "install", "--no-save")
		cmd.Dir = path
		stdoutStderr, err := cmd.CombinedOutput()
		LogToFrontend("info", string(stdoutStderr))
		if err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur npm install dans %s: %v", path, err))
			return err
		}
		LogToFrontend("success", fmt.Sprintf("npm install terminé avec succès dans %s.", path))
	}

	// Parcours récursif des sous-dossiers
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "node_modules" && entry.Name() != ".git" {
			subPath := filepath.Join(path, entry.Name())
			if err := npmInstallRecursive(subPath); err != nil {
				// On log l'erreur mais on continue avec les autres dossiers
				LogToFrontend("warn", fmt.Sprintf("Erreur dans le sous-dossier %s: %v - on continue", subPath, err))
			}
		}
	}
	return nil
}

func TagAction(version, message string) error {
	entries, err := os.ReadDir(".")
	if err != nil {
		LogToFrontend("error", fmt.Sprintf("Erreur lecture répertoire: %v", err))
		return fmt.Errorf("erreur lors de la lecture du répertoire: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		LogToFrontend("info", fmt.Sprintf("TagAction: %s", entry.Name()))
		if err := os.Chdir(entry.Name()); err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur changement de répertoire: %v", err))
			return fmt.Errorf("erreur lors du changement de répertoire: %v", err)
		}

		if _, err := os.Stat("package.json"); err == nil {
			LogToFrontend("info", "package.json existe, on tag")
			if err := execCommand("git", "tag", "-a", version, "-m", message); err != nil {
				LogToFrontend("error", "Erreur git tag")
				return err
			}
			if err := execCommand("git", "push", "--tags"); err != nil {
				LogToFrontend("error", "Erreur git push --tags")
				return err
			}
		}

		if err := os.Chdir(".."); err != nil {
			LogToFrontend("error", fmt.Sprintf("Erreur retour répertoire parent: %v", err))
			return fmt.Errorf("erreur lors du retour au répertoire parent: %v", err)
		}
	}

	return nil
}


