package backend

import (
	"context"

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
