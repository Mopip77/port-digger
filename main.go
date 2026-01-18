package main

import (
	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData())
	systray.SetTitle("Port Digger")
	systray.SetTooltip("Monitor TCP listening ports")

	// Add quit option
	mQuit := systray.AddMenuItem("Quit", "Quit Port Digger")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	// Cleanup if needed
}
