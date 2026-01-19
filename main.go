package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"port-digger/actions"
	"port-digger/llm"
	"port-digger/logger"
	"port-digger/menu"
	"port-digger/scanner"
	"sort"

	"github.com/getlantern/systray"
	"golang.design/x/clipboard"
)

// Global LLM rewriter instance
var rewriter *llm.Rewriter

//go:embed icon/icon.png
var iconData []byte

const version = "1.0.0"

func main() {
	// Initialize logger first
	err := logger.Init()
	if err != nil {
		println("Warning: logger initialization failed:", err.Error())
	}
	defer logger.Close()

	logger.Info("Port Digger v%s starting...", version)

	// Initialize clipboard once at startup
	err = clipboard.Init()
	if err != nil {
		// Non-fatal - clipboard features will just fail silently
		println("Warning: clipboard not available:", err.Error())
		logger.Error("Clipboard initialization failed: %v", err)
	}

	// Initialize LLM rewriter (non-fatal if it fails)
	rewriter, err = llm.NewRewriter()
	if err != nil {
		println("Warning: LLM rewriter not available:", err.Error())
		logger.Error("LLM rewriter initialization failed: %v", err)
	} else {
		logger.Info("LLM rewriter initialized successfully")
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	// Use icon instead of emoji
	systray.SetIcon(iconData)
	systray.SetTooltip(fmt.Sprintf("Port Digger v%s - Monitor TCP Ports", version))

	logger.Info("Building menu...")
	buildMenu()
}

func buildMenu() {
	// Add refresh button
	mRefresh := systray.AddMenuItem("🔄 Refresh", "Rescan ports")
	go func() {
		for range mRefresh.ClickedCh {
			logger.Info("Refresh button clicked, restarting app...")
			// Restart the app to rebuild menu
			restartApp()
		}
	}()

	systray.AddSeparator()

	// Scan ports
	logger.Info("Scanning ports...")
	ports, err := scanner.ScanPorts()
	if err != nil {
		logger.Error("Port scan failed: %v", err)
		systray.AddMenuItem("❌ Scan failed", err.Error())
		addBottomMenu()
		return
	}

	if len(ports) == 0 {
		logger.Info("No ports listening")
		systray.AddMenuItem("No ports listening", "")
		addBottomMenu()
		return
	}

	// Sort by port number
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	logger.Info("Found %d listening ports, adding to menu", len(ports))

	// Add port menu items
	for _, p := range ports {
		addPortMenuItem(p)
	}

	addBottomMenu()
}

// restartApp restarts the application to refresh the menu
func restartApp() {
	executable, err := os.Executable()
	if err != nil {
		logger.Error("Failed to get executable path: %v", err)
		return
	}

	// Start a new instance
	cmd := exec.Command(executable, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		logger.Error("Failed to restart app: %v", err)
		return
	}

	// Exit current instance
	systray.Quit()
}

func addBottomMenu() {
	// Add quit at bottom
	systray.AddSeparator()
	systray.AddMenuItem(fmt.Sprintf("Port Digger v%s", version), "About")

	// LLM Settings submenu
	mLLM := systray.AddMenuItem("⚙️ LLM Settings", "Configure LLM for process name rewriting")
	mLLMOpen := mLLM.AddSubMenuItem("Open Config File", "Edit ~/.config/port-digger/config.yaml")
	mLLMStatus := mLLM.AddSubMenuItemCheckbox("Enabled", "", rewriter != nil && rewriter.IsEnabled())
	mLLMStatus.Disable() // Read-only indicator

	go func() {
		for range mLLMOpen.ClickedCh {
			configPath, err := llm.ConfigPath()
			if err != nil {
				println("Failed to get config path:", err.Error())
				logger.Error("Failed to get config path: %v", err)
				continue
			}
			// Ensure config file exists
			if err := llm.EnsureDefaultConfig(); err != nil {
				println("Failed to create default config:", err.Error())
				logger.Error("Failed to create default config: %v", err)
			}
			// Open in default editor using 'open' command on macOS
			actions.OpenFile(configPath)
		}
	}()

	mQuit := systray.AddMenuItem("Quit", "Quit Port Digger")
	go func() {
		<-mQuit.ClickedCh
		logger.Info("Quit button clicked, exiting...")
		systray.Quit()
	}()
}

// Placeholder for next step
func addPortMenuItem(info scanner.PortInfo) {
	// Get full command for LLM rewriting
	fullCommand := scanner.GetFullCommand(info.PID)
	if fullCommand == "" {
		fullCommand = info.ProcessName
	}

	// Check for cached rewritten name
	var rewrittenName string
	if rewriter != nil && rewriter.IsEnabled() {
		rewrittenName = rewriter.GetServiceName(fullCommand)

		// Trigger async rewrite if not cached
		if rewrittenName == "" {
			rewriter.TriggerRewrite(fullCommand)
		}
	}

	// Format menu item with rewritten name if available
	itemText := menu.FormatPortItemWithRewrite(info, rewrittenName)
	mPort := systray.AddMenuItem(itemText, "")

	// Add submenu items
	mOpen := mPort.AddSubMenuItem("Open in Browser", "Open http://localhost:PORT")
	mCopy := mPort.AddSubMenuItem("Copy Port Number", "Copy to clipboard")
	mPort.AddSubMenuItemCheckbox("------", "", false) // separator-like
	mKill := mPort.AddSubMenuItem(
		fmt.Sprintf("Kill Process (PID: %d)", info.PID),
		"Terminate this process")

	// Handle submenu actions
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				logger.Info("Opening browser for port %d", info.Port)
				actions.OpenBrowser(info.Port)
			case <-mCopy.ClickedCh:
				logger.Info("Copying port %d to clipboard", info.Port)
				err := actions.CopyToClipboard(info.Port)
				if err != nil {
					println("Failed to copy to clipboard:", err.Error())
					logger.Error("Failed to copy port %d to clipboard: %v", info.Port, err)
				}
			case <-mKill.ClickedCh:
				logger.Info("Killing process PID %d (port %d)", info.PID, info.Port)
				err := actions.KillProcess(info.PID)
				if err != nil {
					// Could show notification, but keep it simple for now
					println("Failed to kill process:", err.Error())
					logger.Error("Failed to kill process PID %d: %v", info.PID, err)
				} else {
					logger.Info("Successfully killed process PID %d", info.PID)
					// Restart app to refresh the port list
					logger.Info("Restarting app to refresh port list...")
					restartApp()
				}
			}
		}
	}()
}

func onExit() {
	// Cleanup if needed
}
