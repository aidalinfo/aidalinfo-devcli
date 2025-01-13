package main

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "strings"
)

func submoduleAction(branches ...string) error {
    currentDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("erreur lors de la récupération du répertoire courant: %v", err)
    }
    fmt.Printf(" 👉 On est dans le répertoire %s\n", currentDir)

    fmt.Println(" 🤖 On initialise et update les submodules")
    if err := execCommand("git", "submodule", "init"); err != nil {
        return err
    }
    if err := execCommand("git", "submodule", "update"); err != nil {
        return err
    }

    // Ajouter la branche par défaut à la liste des branches à tenter
    defaultBranch, err := getDefaultBranch()
    if err != nil {
        return fmt.Errorf("erreur lors de la récupération de la branche par défaut : %v", err)
    }
    branches = append(branches, defaultBranch)

    fmt.Printf(" 👉 Branches à essayer : %v\n", branches)

    for _, branch := range branches {
        fmt.Printf(" 🤖 Tentative de checkout de la branche '%s'\n", branch)
        if err := execCommand("git", "checkout", branch); err == nil {
            fmt.Printf(" ✅ Branche '%s' checkoutée avec succès\n", branch)
            break
        }
        fmt.Printf(" ⚠️ Impossible de checkout '%s'\n", branch)
    }

    fmt.Println(" 🤖 On pull")
    if err := execCommand("git", "pull"); err != nil {
        return err
    }

    // Gérer les submodules récursifs
    content, err := os.ReadFile(".gitmodules")
    if err != nil {
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

    fmt.Printf(" 👉 On a trouvé les submodules suivants : %v\n", submodules)

	for _, submodule := range submodules {
		fmt.Printf(" 👉👉 On est dans %s et on a trouvé le submodule: %s\n", currentDir, submodule)
		fmt.Printf(" 🤖 On va dans le répertoire %s\n", submodule)
	
		if err := os.Chdir(submodule); err != nil {
			return fmt.Errorf("erreur lors du changement de répertoire: %v", err)
		}
	
		for _, branch := range branches {
			fmt.Printf(" 🤖 Tentative de checkout de la branche '%s' pour le submodule\n", branch)
			if err := execCommand("git", "checkout", branch); err == nil {
				fmt.Printf(" ✅ Submodule sur branche '%s' checkouté avec succès\n", branch)
				break
			}
			fmt.Printf(" ⚠️ Submodule : Impossible de checkout '%s'\n", branch)
		}
	
		fmt.Println(" 🤖 On pull")
		if err := execCommand("git", "pull"); err != nil {
			return err
		}
	
		// Vérifier s'il y a des submodules récursifs
		if _, err := os.Stat(".gitmodules"); err == nil {
			fmt.Println(" 👉👉 Il y a un fichier .gitmodules")
			fmt.Println(" 🤖🤖 RECURSIVITE !")
			if err := submoduleAction(branches...); err != nil {
				return err
			}
		}
	
		fmt.Printf(" 🤖 On retourne dans le répertoire %s\n", currentDir)
		if err := os.Chdir(currentDir); err != nil {
			return fmt.Errorf("erreur lors du retour au répertoire parent: %v", err)
		}
	}

    return nil
}

func npmAction(all bool) error {
	if !all {
		return nil
	}
    entries, err := os.ReadDir(".")
    if err != nil {
        return fmt.Errorf("erreur lors de la lecture du répertoire: %v", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        fmt.Println(entry.Name())
        if err := os.Chdir(entry.Name()); err != nil {
            return fmt.Errorf("erreur lors du changement de répertoire: %v", err)
        }

        if _, err := os.Stat("package.json"); err == nil {
            fmt.Println("package.json existe, on installe")
            if err := execCommand("npm", "install", "--no-save"); err != nil {
                return err
            }
        }

        if err := os.Chdir(".."); err != nil {
            return fmt.Errorf("erreur lors du retour au répertoire parent: %v", err)
        }
    }

    return nil
}

func tagAction(version, message string) error {
    entries, err := os.ReadDir(".")
    if err != nil {
        return fmt.Errorf("erreur lors de la lecture du répertoire: %v", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        fmt.Println(entry.Name())
        if err := os.Chdir(entry.Name()); err != nil {
            return fmt.Errorf("erreur lors du changement de répertoire: %v", err)
        }

        if _, err := os.Stat("package.json"); err == nil {
            fmt.Println("package.json existe, on tag")
            if err := execCommand("git", "tag", "-a", version, "-m", message); err != nil {
                return err
            }
            if err := execCommand("git", "push", "--tags"); err != nil {
                return err
            }
        }

        if err := os.Chdir(".."); err != nil {
            return fmt.Errorf("erreur lors du retour au répertoire parent: %v", err)
        }
    }

    return nil
}

func execCommand(name string, args ...string) error {
    cmd := exec.Command(name, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
