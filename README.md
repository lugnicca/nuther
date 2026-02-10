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
| All Drives | Summary of all detected drives |
| Details | Detailed S.M.A.R.T. attributes |
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
