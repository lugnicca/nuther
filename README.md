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

### Updates

Nuther can check GitHub Releases and replace the current single-file binary:

```bash
# Check whether a newer release exists
nuther update --check

# Download the matching release asset, verify checksums.txt, and update the binary
nuther update

# Non-interactive/scripted update
nuther update --yes
```

The updater uses the published GoReleaser assets and verifies the release archive SHA-256 from `checksums.txt` before replacing the executable.

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
| `--textfile` | empty | Directory for the Prometheus textfile output; disabled when empty |
| `--report-log` | empty | Path to append a full JSON report line per archived drive; disabled when empty |
| `--once` | `false` | Take one manual snapshot and exit |

On startup, the watcher prints:

- `Event subscription URL`: `GET /events` streams compact JSON events as server-sent events.
- `Snapshot index URL`: `GET /snapshots` lists stored devices and snapshots.
- `GET /snapshots/{id}` returns a specific archived S.M.A.R.T. snapshot.

Snapshots are stored as browsable JSON files under `snapshots/` with an `index.json` alongside them.

### Prometheus textfile

`watch-smart --textfile <dir>` writes `<dir>/nuther-smart.prom` after every
successful scan, for scraping by the node_exporter / Alloy textfile collector:

```bash
sudo ./nuther watch-smart --once --textfile /var/lib/node_exporter/textfile_collector
```

The file is replaced atomically (temp file in the same directory, `0644`, then
rename), so readers never see a partial file. The temp name does not end in
`.prom`, so the collector cannot scrape a half-written file. The output
directory must already exist and be traversable by the collector; it is not
created.

If a run cannot scan any drive, or no drive could be read successfully, it
exits non-zero and leaves the previous file untouched — the file's mtime is the
job-freshness marker. In continuous mode the daemon exits non-zero on shutdown
when `--textfile` is set but the file was never successfully published. A disk
whose own scan errored is still reported, with `nuther_smart_scan_success 0`;
its other series are absent rather than fabricated.

Exposed series (one label set per disk: `device`, `serial`, `model`, `type`,
`interface`):

| Metric | Description |
|--------|-------------|
| `nuther_smart_health_status{status="..."}` | One-hot current status: `GOOD`, `INFO`, `CAUTION`, `BAD` or `UNKNOWN`; other status labels are absent, not `0` |
| `nuther_smart_health_percent` | Remaining drive life (0-100); omitted when the drive health is unknown |
| `nuther_smart_temperature_celsius` | Current temperature |
| `nuther_smart_reallocated_sectors` | Reallocated sectors |
| `nuther_smart_pending_sectors` | Sectors pending reallocation |
| `nuther_smart_uncorrectable_sectors` | Offline uncorrectable sectors |
| `nuther_smart_crc_errors` | UDMA CRC errors (cable/interface) |
| `nuther_smart_power_on_hours` | Power-on hours |
| `nuther_smart_power_cycles` | Power cycles |
| `nuther_smart_wear_leveling_value` | Normalized wear value (omitted when unavailable) |
| `nuther_smart_written_bytes_total` | Total bytes written (omitted when unavailable) |
| `nuther_smart_snapshot_timestamp_seconds` | Timestamp of the snapshot pass |
| `nuther_smart_scan_success` | `1` normally, `0` when that disk's scan errored (only series for such a disk) |

No `host` label is added; Alloy's `external_labels` supplies it.

### JSON report log

`watch-smart --report-log <file>` appends one compact JSON object per line for
each drive that was archived in the scan — the same object the snapshot archive
stores (`id`, `timestamp`, `reason`, `device`, `smart`, with the full ATA
attribute table or NVMe health log):

```bash
sudo ./nuther watch-smart --once --report-log /var/lib/nuther/smart-reports.jsonl
```

With `--once`, every measured drive is archived and then reported; in the
long-running daemon a discovery poll reports only the drives it newly archives
and stops at the first archive or report error, so a poll is not one line per
measured drive.

The whole scan batch is encoded in memory and appended with one `O_APPEND`
write, so a marshal failure leaves nothing behind and a reader cannot see this
batch's lines split by another writer's batch. This is not a per-line atomicity
guarantee: a write may need more than one syscall after a short write, and a
hard failure part-way through can leave a truncated final line. A tailer (for
example Alloy) must ignore any line that does not parse as JSON.

A scan that measured nothing appends nothing and still exits non-zero; a drive
whose own scan errored is not written. The parent directory must already exist.

The file is capped at 8 MiB: once it reaches that size it is renamed to
`<file>.1` (replacing any previous rotation) and a fresh file is started, so
the active path is never rewritten in place and a tailer follows the renamed
file instead of re-reading history. At most two report files exist on disk,
plus the `<file>.lock` sidecar. Rotation only happens while holding a
cross-process lock on that sidecar, so overlapping scans cannot rotate away
each other's active file; if the lock cannot be taken the batch is still
appended unlocked and the run exits non-zero, so the active file may
temporarily exceed the cap. The report file is set to `0644` even under a
restrictive umask, so a non-root collector keeps read access.

### Screenshot directory

The overview screenshot (pressing `s` on the Overview tab) is saved in your home
directory by default, with a filename based on the selected drive's disk id
(device basename) and serial number. To change the save directory, add this to
`~/.config/nuther/config.yaml`:

```yaml
screenshot:
  screenshot_dir: /home/you/Pictures/nuther
```

Use an absolute path (the `~` shortcut is not expanded). The directory can
also be changed from the Settings tab: select **Screenshot Dir**, press
`Enter`, type the path, then `Enter` to commit and `Enter` on any other row
to save.

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
| Snapshots | Archived S.M.A.R.T. snapshot history and known disks |
| Settings | Theme, display options, and screenshot directory |

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
