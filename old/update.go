package main

import (
    "bufio"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
)
var REPO_URL  = "https://github.com/aidalinfo/aidalinfo-devcli"


func promptForUpdate() bool {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Voulez-vous mettre à jour ? (y/n): ")
    response, _ := reader.ReadString('\n')
    response = strings.ToLower(strings.TrimSpace(response))
    return response == "y" || response == "yes"
}

func downloadAndReplace(latestVersion string) error {
    // Construire l'URL de téléchargement en fonction de l'OS et l'architecture
    arch := runtime.GOARCH
    osName := runtime.GOOS
    downloadURL := fmt.Sprintf("%s/releases/download/v%s/aidalinfo-cli_%s_%s", REPO_URL, latestVersion, osName, arch)
    // fmt.Println(downloadURL)
    // Créer un fichier temporaire
    tmpFile, err := os.CreateTemp("", "aidalinfo-devcli")
    if err != nil {
        return fmt.Errorf("erreur lors de la création du fichier temporaire: %v", err)
    }
    defer os.Remove(tmpFile.Name())

    // Télécharger le nouveau binaire
    fmt.Println("Téléchargement de la nouvelle version...")
    resp, err := http.Get(downloadURL)
    if err != nil {
        return fmt.Errorf("erreur lors du téléchargement: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("erreur lors du téléchargement, status: %d", resp.StatusCode)
    }

    // Copier le contenu dans le fichier temporaire
    _, err = io.Copy(tmpFile, resp.Body)
    if err != nil {
        return fmt.Errorf("erreur lors de l'écriture du fichier: %v", err)
    }
    tmpFile.Close()
	fmt.Println("Nom du fichier temporaire :", tmpFile.Name())
    // Rendre le fichier exécutable
    err = os.Chmod(tmpFile.Name(), 0750)
    if err != nil {
        return fmt.Errorf("erreur lors du chmod: %v", err)
    }

    // Obtenir le chemin de l'exécutable actuel
    currentExe, err := os.Executable()
    if err != nil {
        return fmt.Errorf("erreur lors de la récupération du chemin de l'exécutable: %v", err)
    }
    currentExe, err = filepath.EvalSymlinks(currentExe)
    if err != nil {
        return fmt.Errorf("erreur lors de la résolution du symlink: %v", err)
    }

    // Remplacer l'ancien binaire
    fmt.Println("Installation de la nouvelle version...")
    err = os.Rename(tmpFile.Name(), currentExe)
    if err != nil {
        // Si le rename direct échoue, essayer avec sudo
        cmd := exec.Command("sudo", "mv", tmpFile.Name(), currentExe)
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("erreur lors du remplacement du binaire: %v", err)
        }
    }

    fmt.Printf("Mise à jour vers la version %s réussie!\n", latestVersion)
    
    // Relancer l'application avec les mêmes arguments
    args := os.Args[1:]
    cmd := exec.Command(currentExe, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func checkForUpdates(currentVersion string) error {
    fmt.Printf("Vérification des mises à jour ...\n")

    cmd := exec.Command("git", "ls-remote", "--tags", REPO_URL)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("erreur lors de la récupération des tags: %v", err)
    }

    tags := strings.Split(string(output), "\n")
    var latestVersion string
    for _, tag := range tags {
        if strings.Contains(tag, "refs/tags/") && !strings.Contains(tag, "^{}") {
            parts := strings.Split(tag, "refs/tags/")
            if len(parts) == 2 {
                latestVersion = strings.TrimSpace(parts[1])
            }
        }
    }

    if latestVersion == "" {
        return fmt.Errorf("aucune version trouvée")
    }

    latestVersion = strings.TrimPrefix(latestVersion, "v")
    currentVersion = strings.TrimPrefix(currentVersion, "v")

    if latestVersion != currentVersion {
        fmt.Printf("Une nouvelle version est disponible : %s (version actuelle : %s)\n", latestVersion, currentVersion)
        
        if promptForUpdate() {
            return downloadAndReplace(latestVersion)
        }
        fmt.Println("Mise à jour annulée par l'utilisateur")
    } else {
        fmt.Println("Vous utilisez la dernière version disponible.")
    }

    return nil
}