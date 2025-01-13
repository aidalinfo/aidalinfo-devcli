package main

import (
    "fmt"
    "flag"
    // "regexp"
)

var VERSION = "0.1.0"

func main() {
    projectPath := flag.String("path", ".", "Chemin du projet")
    uiMode := flag.Bool("ui", false, "Lancer l'interface utilisateur")
    version := flag.Bool("version", false, "Afficher la version")
    v := flag.Bool("v", false, "Afficher la version")
	flag.Parse()
    // Appelle listSubmodule pour le répertoire courant
    // submodules, err := listSubmodule("/home/killian/dev/aidalinfo/PROJET-pulse-myIT")


    // // Affiche les sous-modules trouvés
    // fmt.Println("Sous-modules trouvés :")
    // for _, submodule := range submodules {
    //     fmt.Println(submodule)
    // }


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
    } else {
		fmt.Println("Usage: --ui pour lancer l'interface utilisateur && --path pour spécifier le chemin du projet")
	}
}