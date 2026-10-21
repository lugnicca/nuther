# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.0] - unreleased

### Added

- `nuther watch-smart --report-log <file>` appends one compact JSON line per
  drive archived in a scan — the same object the snapshot archive stores
  (`id`, `timestamp`, `reason`, `device`, `smart`, including the full ATA
  attribute table or NVMe health log). With `--once` this is every measured
  drive; in the long-running daemon the discovery polls report only the drives
  they newly archive. The whole scan batch is one in-memory buffer written
  with a single `O_APPEND` write, so a marshal failure writes nothing and a
  tailer cannot see the batch split by another writer. A hard failure
  part-way through the write can still leave a truncated final line, so a
  tailer must ignore a line that does not parse as JSON. A scan that measured
  nothing appends nothing and still exits non-zero. The file is capped at
  8 MiB: once it reaches the cap it is renamed to `<file>.1` and a fresh file
  is started, so the active path is never rewritten in place and a tailer does
  not re-read history. Rotation and append are serialized by a cross-process
  lock, and the file is set to 0644 so a non-root tailer keeps read access.

## [0.6.0] - 2026-10-21

### Added

- `nuther watch-smart --textfile <dir>` writes a Prometheus textfile
  (`nuther-smart.prom`) after every successful scan, atomically, for the
  node_exporter / Alloy textfile collector. The output directory must already
  exist. When nothing can be scanned, or no drive can be read successfully, the
  run exits non-zero and the previous file is left untouched; in continuous
  mode the daemon exits non-zero on shutdown when the file was never
  successfully published. A disk whose own scan errored is still reported with
  `nuther_smart_scan_success 0` and no other series.

## [0.1.0] - 2025-01-01

### Added

- S.M.A.R.T. disk health monitoring via smartmontools
- Support for SATA HDD, SATA SSD, and NVMe drives
- Health status indicators: GOOD, CAUTION, BAD
- Temperature monitoring with visual gauges
- Overview tab with drive info and key attributes
- All Drives tab with multi-drive summary
- Details tab with full S.M.A.R.T. attribute table
- Settings tab with theme and configuration options
- 11 built-in color themes (default, atomic-tangerine, catppuccin, dracula, everforest, gruvbox, nord, rose-petale, solarized-dark, sous-bois, tokyo-night)
- Custom theme support via YAML files
- Vim-style keyboard navigation
- Screenshot capture to clipboard
- Demo mode for running without root privileges
- Cross-platform support (Linux, macOS, Windows)
- Configurable keybindings via YAML
- Celsius/Fahrenheit temperature display toggle
