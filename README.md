# Nuther

A terminal-based S.M.A.R.T. disk health monitor inspired by CrystalDiskInfo.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=flat)

```
  ███╗   ██╗██╗   ██╗████████╗██╗  ██╗███████╗██████╗
  ████╗  ██║██║   ██║╚══██╔══╝██║  ██║██╔════╝██╔══██╗
  ██╔██╗ ██║██║   ██║   ██║   ███████║█████╗  ██████╔╝
  ██║╚██╗██║██║   ██║   ██║   ██╔══██║██╔══╝  ██╔══██╗
  ██║ ╚████║╚██████╔╝   ██║   ██║  ██║███████╗██║  ██║
  ╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
```

## Features

- Automatic detection of SATA HDDs, SATA SSDs, and NVMe drives
- Complete S.M.A.R.T. data with health indicators
- Temperature monitoring with visual gauges
- Health alerts: GOOD / CAUTION / BAD status
- Multiple themes support
- Vim-style keyboard navigation
- Screenshot capture to clipboard
- Cross-platform (Linux, macOS, Windows)

## Requirements

- [smartmontools](https://www.smartmontools.org/) must be installed
- Root/Administrator privileges for full S.M.A.R.T. access

```bash
# Debian/Ubuntu
sudo apt install smartmontools

# Fedora/RHEL
sudo dnf install smartmontools

# Arch Linux
sudo pacman -S smartmontools

# macOS
brew install smartmontools

# Windows
# Download from https://www.smartmontools.org/wiki/Download
```

## Installation

### From source

```bash
git clone https://github.com/lugnicca/nuther.git
cd nuther
go build ./cmd/nuther
```

### Build options

```bash
# Simple build
go build ./cmd/nuther

# With version info
go build -ldflags "-X main.version=1.0.0" ./cmd/nuther

# Cross-compilation
GOOS=linux GOARCH=amd64 go build ./cmd/nuther
GOOS=darwin GOARCH=arm64 go build ./cmd/nuther
GOOS=windows GOARCH=amd64 go build ./cmd/nuther
```

## Usage

```bash
# Linux/macOS - run with sudo for full access
sudo ./nuther

# Windows - run as Administrator
nuther.exe

# Demo mode (no privileges required)
./nuther
```

### S.M.A.R.T. snapshot watcher

Run the lightweight watcher to archive S.M.A.R.T. reports when drives appear and, optionally, on a fixed schedule:

```bash
# Watch for newly connected drives and expose local event/snapshot HTTP routes
sudo ./nuther watch-smart

# Take one manual snapshot of currently detected drives and exit
sudo ./nuther watch-smart --once

# Poll every minute and snapshot every hour
sudo ./nuther watch-smart --interval 1m --snapshot-interval 1h

# Send each snapshot event to n8n
sudo ./nuther watch-smart --webhook-url https://n8n.example/webhook/smart
```

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--interval` | `30s` | Device detection polling interval |
| `--snapshot-interval` | `0` | Periodic snapshot interval; disabled when `0` |
| `--store` | `~/.config/nuther/smart-snapshots` | JSON snapshot archive directory |
| `--listen` | `127.0.0.1:0` | Local HTTP address for event and snapshot routes |
| `--webhook-url` | empty | Optional outbound webhook URL for snapshot events |
| `--once` | `false` | Take one manual snapshot and exit |

On startup, the watcher prints:

- `Event subscription URL`: `GET /events` streams compact JSON events as server-sent events.
- `Snapshot index URL`: `GET /snapshots` lists stored devices and snapshots.
- `GET /snapshots/{id}` returns a specific archived S.M.A.R.T. snapshot.

Snapshots are stored as browsable JSON files under `snapshots/` with an `index.json` alongside them.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch tabs |
| `n` / `]` | Next drive |
| `p` / `[` | Previous drive |
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `r` | Refresh drive data |
| `s` | Screenshot to clipboard |
| `?` | Toggle help |
| `q` | Quit |

## Tabs

| Tab | Description |
|-----|-------------|
| Overview | Drive info, health status, temperature gauge |
| S.M.A.R.T. Attributes | Detailed S.M.A.R.T. attribute table |
| Details | Detailed S.M.A.R.T. attributes |
| All Drives | Summary of all detected drives |
| Sector Grid | Square cell grid for the selected disk sector pressure |
| Settings | Theme and configuration |

## Health Status

| Status | Meaning |
|--------|---------|
| GOOD | Drive is healthy |
| CAUTION | Monitoring recommended |
| BAD | Action required |

## Development

### Project Structure

```
nuther/
├── cmd/nuther/         # Entry point
├── internal/
│   ├── config/         # Configuration and themes
│   ├── platform/       # OS-specific code
│   ├── screenshot/     # Screenshot capture
│   ├── smart/          # S.M.A.R.T. data handling
│   └── ui/             # Terminal UI (Bubble Tea)
│       ├── components/ # Reusable UI components
│       ├── views/      # Tab views
│       └── styles/     # Styling
└── testdata/           # Test fixtures
```

### Running Tests

```bash
# All tests
go test ./...

# With verbose output
go test ./... -v

# Specific package
go test ./internal/smart/... -v

# With coverage
go test ./... -cover
```

## Credits

- Inspired by [CrystalDiskInfo](https://crystalmark.info/en/software/crystaldiskinfo/)
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Uses [smartmontools](https://www.smartmontools.org/) for S.M.A.R.T. data

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
