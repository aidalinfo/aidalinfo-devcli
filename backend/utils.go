package backend

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var AppCtxForLogToFrontend context.Context // à setter dans app.go

func LogToFrontend(level, msg string) {
	if AppCtxForLogToFrontend == nil {
		return
	}
	prefix := ""
	switch level {
	case "debug":
		prefix = "[DEBUG]"
	case "info":
		prefix = "[INFO]"
	case "warn":
		prefix = "[WARN]"
	case "error":
		prefix = "[ERROR]"
	case "success":
		prefix = "[SUCCESS]"
	}
	fullMsg := prefix + " " + msg
	println(fullMsg) // Affiche dans la console du backend
	runtime.EventsEmit(AppCtxForLogToFrontend, "backend-log", fullMsg)
}

// execCommand exécute une commande et retourne une erreur si elle échoue
func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	
	// Créer un pipe pour capturer stdout et stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Démarrer la commande
	if err := cmd.Start(); err != nil {
		return err
	}

	// Lire stdout en temps réel
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			LogToFrontend("info", scanner.Text())
			fmt.Println(scanner.Text()) // Afficher aussi dans le terminal
		}
	}()

	// Lire stderr en temps réel
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			LogToFrontend("warning", scanner.Text())
			fmt.Println(scanner.Text()) // Afficher aussi dans le terminal
		}
	}()

	// Attendre la fin de la commande
	return cmd.Wait()
}


// execCommandOutput exécute une commande et retourne sa sortie
func execCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
