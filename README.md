# Port Digger

A lightweight macOS menu bar tool for monitoring TCP listening ports.

## Features

- 🔍 Real-time port monitoring (on-demand, no background polling)
- 🔄 Refresh button to rescan ports (automatically restarts the app)
- 🌐 Open ports in browser with one click
- 📋 Copy port numbers to clipboard
- ⚡ Kill processes (with sudo prompt when needed, auto-refreshes after kill)
- 🤖 LLM-powered process name rewriting (optional)
- 📝 Comprehensive logging for debugging
- 💾 Minimal memory footprint (~10-20MB)

## Screenshot

```
 3000 • node
 8080 • Python
27017 • mongod
```

**Example Actions:**
- Click "3000 • node" → See submenu
  - Open in Browser → Opens http://localhost:3000
  - Copy Port Number → "3000" in clipboard
  - Kill Process (PID: 12345) → Prompts for password, terminates node

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/yourusername/port-digger.git
cd port-digger

# Build
make build

# Run
./PortDigger
```

### Manual Install

```bash
make install
# Starts PortDigger from ~/Applications/
```

## Usage

1. Click the menu bar icon to see all listening TCP ports
2. Ports are sorted by number and show process name
3. Hover over any port to see actions:
   - **Open in Browser** - Opens `http://localhost:PORT`
   - **Copy Port Number** - Copies port to clipboard
   - **Kill Process** - Terminates the process (asks for password if needed, auto-refreshes)
4. Click **Refresh** to rescan ports (restarts the app to get fresh data)

## Logging

Port Digger automatically logs all operations to help with debugging:

- **Log Location**: `~/.config/port-digger/logs/port-digger.log`
- **What's Logged**:
  - Application startup and shutdown
  - Port scanning operations (lsof commands and results)
  - LLM API requests and responses (if enabled)
  - User actions (opening browser, copying to clipboard, killing processes)
  - Errors and warnings

**View logs:**
```bash
# View all logs
cat ~/.config/port-digger/logs/port-digger.log

# Follow logs in real-time
tail -f ~/.config/port-digger/logs/port-digger.log

# View recent logs
tail -n 50 ~/.config/port-digger/logs/port-digger.log
```

## LLM Integration

Port Digger can use LLM to rewrite process names for better readability:

1. Open LLM Settings from the menu bar
2. Edit the config file at `~/.config/port-digger/config.yaml`
3. Configure your LLM API endpoint and key
4. Enable the feature and restart the app

**Example**: `node /opt/homebrew/bin/claude-code-ui` → `claude-code-ui ✨`

## Requirements

- macOS 10.13+
- Go 1.21+ (for building from source)

## Technical Details

- **Runtime Memory**: 10-20MB
- **Binary Size**: 8-15MB
- **Dependencies**: systray, open-golang, clipboard
- **Scan Method**: `lsof -iTCP -sTCP:LISTEN -nP`

## Testing

```bash
make test
```

## License

MIT
