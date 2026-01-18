package main

import (
	"sort"
	"github.com/getlantern/systray"
	"port-digger/scanner"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData())
	systray.SetTitle("Port Digger")
	systray.SetTooltip("Monitor TCP listening ports")

	refreshMenu()
}

func refreshMenu() {
	// Clear menu - systray doesn't have ResetMenu, we'll work around this
	// by just building the menu once and updating it dynamically
	// For now, let's just build the menu structure

	// Add refresh button
	mRefresh := systray.AddMenuItem("🔄 Refresh", "Rescan ports")
	go func() {
		for range mRefresh.ClickedCh {
			// Note: systray doesn't support dynamic menu rebuild easily
			// This is a limitation we'll note
		}
	}()

	systray.AddSeparator()

	// Scan ports
	ports, err := scanner.ScanPorts()
	if err != nil {
		systray.AddMenuItem("❌ Scan failed", err.Error())
		return
	}

	if len(ports) == 0 {
		systray.AddMenuItem("No ports listening", "")
		return
	}

	// Sort by port number
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	// Add port menu items (implementation in next step)
	for _, p := range ports {
		addPortMenuItem(p)
	}

	// Add quit at bottom
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit Port Digger")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

// Placeholder for next step
func addPortMenuItem(info scanner.PortInfo) {
	// TODO: implement in next task
}

func onExit() {
	// Cleanup if needed
}
