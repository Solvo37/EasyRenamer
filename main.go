package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:       "EasyRenamer",
		Width:       1600,
		Height:      960,
		MinWidth:    1180,
		MinHeight:   720,
		Frameless:   true,
		BackgroundColour: options.NewRGB(10, 18, 30),
		AssetServer: &assetserver.Options{Assets: assets},
		OnStartup:   app.startup,
		OnShutdown:  app.shutdown,
		Bind: []interface{}{app},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		Windows: &windows.Options{
			Theme:                 windows.Dark,
			WebviewIsTransparent:  false,
			WindowIsTranslucent:   false,
			BackdropType:          windows.None,
			DisablePinchZoom:      true,
			IsZoomControlEnabled:  false,
			ResizeDebounceMS:      4,
		},
	})
	if err != nil {
		println("EasyRenamer:", err.Error())
	}
}
