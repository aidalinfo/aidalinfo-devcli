package main

import (
	"aidalinfo-copilot/cmd"
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Check if running in CLI mode (if there are command line arguments)
	if len(os.Args) > 1 {
		// Run in CLI mode with Cobra
		cmd.Execute()
		return
	}

	// Otherwise, run in GUI mode with Wails
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "aidalinfo-copilot",
		Width:            1920,
		Height:           1080,
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
