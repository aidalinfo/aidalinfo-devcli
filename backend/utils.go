package backend

import (
	"context"
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		LogToFrontend("error", string(output))
		return err
	}
	LogToFrontend("debug", string(output))
	return nil
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
