package main

import (
    "fmt"
    "flag"
    // "regexp"
)

func main() {
    projectPath := flag.String("path", ".", "Chemin du projet")
    uiMode := flag.Bool("ui", false, "Lancer l'interface utilisateur")
	flag.Parse()
    // Appelle listSubmodule pour le répertoire courant
    // submodules, err := listSubmodule("/home/killian/dev/aidalinfo/PROJET-pulse-myIT")
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

    // Affiche les sous-modules trouvés
    fmt.Println("Sous-modules trouvés :")
    for _, submodule := range submodules {
        fmt.Println(submodule)
    }



	if *uiMode {
		// Lancer l'interface utilisateur
		RunUI(submodules, submoduleNames)
	} else {
		fmt.Println("Usage: --ui pour lancer l'interface utilisateur && --path pour spécifier le chemin du projet")
	}
}