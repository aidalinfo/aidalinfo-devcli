package main

import (
	"flag"
	"fmt"
	"os"

	// "regexp"
	"strings"
)

var VERSION = "0.0.6"

func main() {

	listTest()
	projectPath := flag.String("path", ".", "Chemin du projet")
	uiMode := flag.Bool("ui", false, "Lancer l'interface utilisateur")
	version := flag.Bool("version", false, "Afficher la version")
	v := flag.Bool("v", false, "Afficher la version")
	//Setup args
	installCmd := flag.Bool("install", false, "Installer les submodules")
	branchArg := flag.String("branch", "", "Spécifier la ou les branches (séparées par un espace)")
	npmCmd := flag.Bool("npm", false, "Installer les dépendances npm")
	fullCmd := flag.Bool("full", false, "Installation complète (submodules + npm)")
	devopsCmd := flag.Bool("ui-devops", false, "Lancer l'interface DevOps")
	flag.Parse()

	if err := checkForUpdates(VERSION); err != nil {
		fmt.Println("Erreur lors de la vérification des mises à jour:", err)
	}

	// Traitement des commandes d'installation
	if *installCmd {
		if *branchArg != "" {
			// Diviser l'argument branch en cas de branches multiples
			branches := strings.Fields(*branchArg)
			fmt.Printf("Installer les sous-modules avec les branches : %s\n", branches)
			if err := submoduleAction(branches...); err != nil {
				fmt.Printf("Erreur: %v\n", err)
				os.Exit(1)
			}
			if *npmCmd {
				if err := npmAction(true); err != nil {
					fmt.Printf("Erreur: %v\n", err)
					os.Exit(1)
				}
			}
		} else if *fullCmd {
			if err := submoduleAction(); err != nil {
				fmt.Printf("Erreur: %v\n", err)
				os.Exit(1)
			}
			if err := npmAction(true); err != nil {
				fmt.Printf("Erreur: %v\n", err)
				os.Exit(1)
			}
		} else if *npmCmd {
			if err := npmAction(true); err != nil {
				fmt.Printf("Erreur: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := submoduleAction(); err != nil {
				fmt.Printf("Erreur: %v\n", err)
				os.Exit(1)
			}
		}
		return
	}
	if *uiMode {
		submodules, err := listSubmodule(*projectPath)
		if err != nil {
			fmt.Println("Erreur :", err)
			return
		}
		// Appelle cleanSubmodule pour nettoyer les chemins
		submoduleNames, err := cleanSubmodule(submodules)
		if err != nil {
			fmt.Println("Erreur :", err)
			return
		}
		// Lancer l'interface utilisateur
		RunUI(submodules, submoduleNames)
	} else if *version || *v {
		fmt.Println("Aidalinfo devcli version", VERSION)
	} else if *devopsCmd {
		// Appelle cleanSubmodule pour nettoyer les chemins
		submodules, err := listSubmodule(*projectPath)
		submoduleNames, err := cleanSubmodule(submodules)
		if err != nil {
			fmt.Println("Erreur :", err)
			return
		}
		RunDevOpsUI(submodules, submoduleNames)
	} else {
		fmt.Println("Usage:")
		fmt.Println("  -ui              Lancer l'interface utilisateur")
		fmt.Println("  -ui-devops       Lancer l'interface DevOps (only tags pour le moment)")
		fmt.Println("  -path            Spécifier le chemin du projet")
		fmt.Println("  -install         Installer les submodules")
		fmt.Println("  -branch=\"X Y\"    Spécifier la ou les branches (X avec fallback sur Y)")
		fmt.Println("  -npm             Installer les dépendances npm")
		fmt.Println("  -full            Installation complète (submodules + npm)")
		fmt.Println("  -version, -v     Afficher la version")
	}
}
