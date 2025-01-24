package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// listSubmodule : Récupère les sous-modules récursivement et retourne leurs chemins
func listSubmodule(path string) ([]string, error) {
	var results []string

	// Si le chemin est vide, utilise le répertoire courant
	if path == "" {
		path = "."
	}
	// fmt.Println(path)
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

func cleanSubmodule(submodules []string) ([]string, error) {
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

// Fonction pour effectuer un merge
func createMerge(currentBranch, targetBranch, repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "merge", "--no-ff", targetBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Erreur lors du merge : %s\n%s", err.Error(), string(output))
	}
	// Effectuer le push
	pushCmd := exec.Command("git", "-C", repoPath, "push", "origin", currentBranch)
	pushOutput, pushErr := pushCmd.CombinedOutput()
	if pushErr != nil {
		return fmt.Errorf("Erreur lors du push : %s\n%s", pushErr.Error(), string(pushOutput))
	}
	return nil
}

func getDiffSummary(currentBranch, targetBranch, repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "diff", "--shortstat", currentBranch+".."+targetBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Erreur lors de l'obtention des différences : %s\n%s", err.Error(), string(output))
	}
	diffSummary := strings.TrimSpace(string(output))
	if diffSummary == "" {
		return "Aucune différence entre les deux branches", nil
	}
	return diffSummary, nil
}

// Fonction pour effectuer un push
func pushChanges(currentBranch, repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "push", "origin", currentBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Erreur lors du push : %s\n%s", err.Error(), string(output))
	}
	return nil
}

type Commit struct {
	Date      string
	Author    string
	Message   string
	Submodule string
}

func getLastCommits(submodules []string) ([]Commit, error) {
	var allCommits []Commit

	for _, submodule := range submodules {
		// On garde le hash dans le format uniquement pour le tri, mais on ne le stocke pas
		cmd := exec.Command("git", "-C", submodule, "log", "--all", "-n", "3", "--pretty=format:%ai|%an|%s")
		output, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}

		commits := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, commit := range commits {
			if commit == "" {
				continue
			}
			parts := strings.Split(commit, "|")
			if len(parts) == 3 {
				allCommits = append(allCommits, Commit{
					Date:      parts[0],
					Author:    parts[1],
					Message:   parts[2],
					Submodule: filepath.Base(submodule),
				})
			}
		}
	}

	// Trier les commits par date (du plus récent au plus ancien)
	sort.Slice(allCommits, func(i, j int) bool {
		return allCommits[i].Date > allCommits[j].Date
	})

	// Retourner les 20 commits les plus récents
	if len(allCommits) > 20 {
		return allCommits[:20], nil
	}
	return allCommits, nil
}

// Fonction pour récupérer la branche par défaut de GitHub
func getDefaultBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("impossible de déterminer la branche par défaut : %v", err)
	}

	// Extraire la branche par défaut du chemin
	defaultBranch := strings.TrimSpace(strings.TrimPrefix(string(output), "refs/remotes/origin/"))
	fmt.Printf(" 👉 Branche par défaut détectée : %s\n", defaultBranch)
	return defaultBranch, nil
}

// Fonction qui récupère d'un repos
func getLastTags(repoPath string) ([]string, []string, error) {
	// Obtenir tous les tags triés par date
	cmd := exec.Command("git", "-C", repoPath, "for-each-ref", "--sort=-creatordate", "--format=%(refname:short)", "refs/tags/")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("Erreur lors de la récupération des tags : %s\n%s", err.Error(), string(output))
	}

	// Séparer les tags en `v*` et `rc-*`
	tags := strings.Split(string(output), "\n")
	var vTags []string
	var rcTags []string

	for _, tag := range tags {
		if strings.HasPrefix(tag, "v") {
			vTags = append(vTags, tag)
		} else if strings.HasPrefix(tag, "rc-") {
			rcTags = append(rcTags, tag)
		}
	}

	return vTags, rcTags, nil
}

// Creation d'un tag
func createTag(repoPath, version, message string) error {
	cmd := exec.Command("git", "-C", repoPath, "tag", "-a", version, "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Erreur : %s\n%s", err, string(output))
	}
	pushCmd := exec.Command("git", "-C", repoPath, "push", "--tags")
	output, err = pushCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Erreur lors du push des tags : %s\n%s", err, string(output))
	}
	return nil
}

// Fonction pour récupérer les modifications en attente
func getWaitingChanges(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Erreur lors de la récupération des modifications : %s\n%s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func changeBranche(repoPath, branch string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Erreur lors du changement de branche : %s\n%s", err, string(output))
	}
	return nil
}

func gitUpdateAction(submodules []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
			return fmt.Errorf("erreur lors de la récupération du répertoire courant: %v", err)
	}

	for _, submodule := range submodules {
			fmt.Printf("📦 Mise à jour git dans %s\n", submodule)
			if err := os.Chdir(submodule); err != nil {
					return fmt.Errorf("erreur lors du changement de répertoire: %v", err)
			}

			if err := execCommand("git", "pull"); err != nil {
					return err
			}

			if err := os.Chdir(currentDir); err != nil {
					return fmt.Errorf("erreur lors du retour au répertoire parent: %v", err)
			}
	}

	return nil
}