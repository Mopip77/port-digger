package main

import (
	_ "embed"
	"fmt"
	"port-digger/actions"
	"port-digger/menu"
	"port-digger/scanner"
	"sort"

	"github.com/getlantern/systray"
	"golang.design/x/clipboard"
)

//go:embed icon/icon.png
var iconData []byte

const version = "1.0.0"

func main() {
	// Initialize clipboard once at startup
	err := clipboard.Init()
	if err != nil {
		// Non-fatal - clipboard features will just fail silently
		println("Warning: clipboard not available:", err.Error())
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	// Use icon instead of emoji
	systray.SetIcon(iconData)
	systray.SetTooltip(fmt.Sprintf("Port Digger v%s - Monitor TCP Ports", version))

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
	systray.AddMenuItem(fmt.Sprintf("Port Digger v%s", version), "About")
	mQuit := systray.AddMenuItem("Quit", "Quit Port Digger")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

// Placeholder for next step
func addPortMenuItem(info scanner.PortInfo) {
	// Format: " 3000 • node"
	itemText := menu.FormatPortItem(info)
	mPort := systray.AddMenuItem(itemText, fmt.Sprintf("Port %d - %s (PID: %d)",
		info.Port, info.ProcessName, info.PID))

	// Add submenu items
	mOpen := mPort.AddSubMenuItem("Open in Browser", "Open http://localhost:PORT")
	mCopy := mPort.AddSubMenuItem("Copy Port Number", "Copy to clipboard")
	mPort.AddSubMenuItemCheckbox("", "", false) // separator-like
	mKill := mPort.AddSubMenuItem(
		fmt.Sprintf("Kill Process (PID: %d)", info.PID),
		"Terminate this process")

	// Handle submenu actions
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				actions.OpenBrowser(info.Port)
			case <-mCopy.ClickedCh:
				err := actions.CopyToClipboard(info.Port)
				if err != nil {
					println("Failed to copy to clipboard:", err.Error())
				}
			case <-mKill.ClickedCh:
				err := actions.KillProcess(info.PID)
				if err != nil {
					// Could show notification, but keep it simple for now
					println("Failed to kill process:", err.Error())
				}
			}
		}
	}()
}

func onExit() {
	// Cleanup if needed
}
