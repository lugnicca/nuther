# Nuther Capacity Test — Design Specification

**Status:** Approved design  
**Date:** 2026-07-25  
**Target release:** Nuther V1 capacity-testing feature  
**Platforms:** Linux, macOS, Windows

## 1. Purpose

Add a native, cross-platform capacity and sustained-performance test to Nuther. The feature must verify that a storage device can store and return the amount of data it claims, detect fake-capacity behavior and corruption, measure write/read performance over time, protect devices from excessive temperature, and produce a resumable, auditable report.

The feature supports two execution modes:

1. **Safe file mode:** fills configurable free space using dedicated test files without overwriting existing data.
2. **Destructive raw-device mode:** writes directly to a selected device or range after strict safety checks and explicit confirmation.

The implementation must use a common native Go engine rather than depend on F3, H2testw, or another external benchmark. It may reproduce the underlying validation method: deterministic data tied to logical offsets is written, read back, and verified.

## 2. Goals

- Validate actual writable and readable capacity.
- Detect data corruption, duplicated regions, address wraparound, missing data, short reads, and I/O failures.
- Report real performance as a time series, not only an average.
- Distinguish application write throughput from durably synchronized throughput.
- Support controlled full-device runs lasting many hours.
- Support fixed batch cooldowns and adaptive thermal throttling.
- Resume safely after process termination, reboot, crash, or device disconnection.
- Provide equivalent semantics and report structures on Linux, macOS, and Windows.
- Integrate into the Nuther TUI while exposing the same engine through a CLI.
- Export a versioned, exhaustive JSON report.

## 3. Non-goals for V1

- HTML, CSV, PNG, or hosted report export.
- Cloud synchronization or remote execution.
- Replacing a general filesystem benchmark suite.
- Claiming full-device validation from a sampled or partial run.
- Destructive testing of the currently active system disk.
- Automatic repair, formatting, partitioning, or recovery of a failed device.
- Mandatory integration with external tools such as F3 or H2testw.

## 4. Architectural Decision

Use a **common orchestration core with specialized storage backends**.

A single engine owns:

- run state and transitions;
- deterministic pattern generation and verification;
- batching and concurrency;
- checkpoints and safe resume;
- thermal regulation;
- temporal metrics and statistics;
- error classification;
- report generation.

Backends own only physical storage operations:

- open and close;
- read and write at logical offsets;
- synchronize or flush;
- expose usable size and alignment constraints;
- report supported I/O modes.

Two backend families are required:

- **File backend:** creates numbered test files in a dedicated directory on a mounted filesystem.
- **Raw-device backend:** performs aligned direct device access through platform-specific adapters.

This structure avoids duplicating validation, metrics, thermal logic, and checkpoint behavior while preserving the distinct guarantees and safety requirements of file and raw testing.

## 5. Proposed Package Boundaries

```text
internal/capacitytest/
├── engine/          orchestration and persistent state machine
├── pattern/         deterministic generation and verification
├── backend/
│   ├── file/        non-destructive file-based operations
│   └── raw/         destructive raw-device operations
├── platform/        Linux, macOS, and Windows adapters
├── thermal/         sensing, throttling, cooldown, abort policy
├── metrics/         temporal samples and aggregate statistics
├── checkpoint/      atomic manifests and safe resume
├── safety/          preflight checks and system-disk protection
└── report/          versioned report model and JSON encoding
```

Each package must expose a small interface that can be tested independently. Platform-specific APIs must not leak into the engine.

## 6. Run Lifecycle

Every run is a persisted state machine.

### 6.1 Preflight

Before any write, Nuther must:

- resolve a stable device identity;
- record model, serial, capacity, transport, logical sector size, and physical sector size when available;
- detect whether the target hosts the running operating system;
- determine privileges and backend capabilities;
- detect mounted partitions, open volumes, swap/paging use, RAID membership, and dependent volumes where supported;
- calculate the exact target range and expected write volume;
- validate alignment, block size, concurrency, cache, and synchronization settings;
- probe temperature telemetry availability;
- estimate duration as a range rather than a falsely precise timestamp;
- estimate device write impact, especially for SSDs;
- persist the immutable run configuration before starting.

A failed preflight performs no test write.

### 6.2 Write

The engine writes deterministic content associated with:

- report format and pattern version;
- run identifier;
- stable target identity;
- logical offset;
- optional user-selected seed.

This allows verification to distinguish corruption from data returned from the wrong logical address.

The write phase supports continuous execution or configured batches. Metrics are sampled periodically throughout the run.

### 6.3 Cooldown and Thermal Regulation

Between batches, the run may pause for a fixed duration. When temperature telemetry is available, adaptive regulation may additionally reduce load, pause, resume, or abort.

### 6.4 Verify

The engine rereads the configured range and verifies content against the expected deterministic pattern. It records:

- valid bytes and ranges;
- first failing offset;
- corrupted ranges;
- duplicated or incorrectly addressed data;
- short reads;
- timeouts and operating-system I/O errors;
- unusually slow windows.

### 6.5 Finalize

The engine:

- flushes and closes the backend where possible;
- atomically saves the final state and report;
- cleans up file-mode artifacts according to policy;
- retains the run manifest and JSON report;
- assigns a typed verdict.

Required final verdicts include:

- `PASS`
- `PARTIAL_PASS`
- `CAPACITY_MISMATCH`
- `DATA_CORRUPTION`
- `IO_FAILURE`
- `THERMAL_ABORT`
- `SAFETY_ABORT`
- `USER_ABORT`
- `INCOMPLETE`

`PASS` is permitted only when the complete intended range was written, durably synchronized according to the selected policy, read, and verified. A sampled test must never produce a full-capacity `PASS`.

## 7. Safe File Mode

File mode operates in a Nuther-owned directory on a mounted filesystem.

### 7.1 Behavior

- Measure available space before the run.
- Reserve a safety margin, defaulting to 5% with a non-zero minimum.
- Create numbered files with deterministic contents.
- Split files and batches according to configuration and platform constraints.
- Synchronize according to the selected policy.
- Reread and verify every completed file.
- Delete files after successful verification by default.
- Permit explicit retention for diagnosis.

### 7.2 Guarantees

File mode validates only the space actually written and verified. Reports must separately show:

- reported volume capacity;
- space occupied before the run;
- free space before the run;
- bytes written;
- bytes synchronized;
- bytes verified;
- percentage of total volume capacity validated.

Nuther must not describe free-space validation as whole-device validation.

### 7.3 Filesystem Safety

- On the active system volume, enforce a minimum free-space reserve that cannot be disabled.
- Refuse configurations that would cross the enforced reserve.
- Detect an unexpected fall in free space and stop safely before exhausting the volume where feasible.
- Use a recognizable Nuther directory and manifest so stale files can be attributed and safely recovered.

## 8. Destructive Raw-device Mode

Raw mode directly overwrites the selected device or selected range.

### 8.1 Mandatory Preconditions

- Administrative/root privileges.
- Target is not the active system disk.
- Target and dependent volumes are unmounted or taken offline.
- Target is not used for swap, paging, active RAID, or another live dependency.
- Stable identity remains unchanged from preflight to first write.
- The user manually enters the stable identifier shown by Nuther.
- The final confirmation clearly states that the selected range will be destroyed.

The operating-system device path alone is never considered a stable identity.

### 8.2 Active System Disk

Destructive testing of the disk hosting the currently running system is refused absolutely. Testing such a disk requires booting another environment where it is no longer the active system disk.

### 8.3 Non-interactive Execution

A generic `--yes` flag must not be sufficient. Automation requires all of:

- an explicit destructive flag;
- expected stable identifier;
- expected capacity or a bounded tolerance;
- exact intended range;
- successful safety preflight.

A mismatch aborts before the first write.

## 9. Range and Test Coverage

Expert mode supports:

- complete target;
- exact start and end offsets;
- initial, middle, or final percentage ranges;
- distributed samples;
- continuation of an interrupted configured range.

The report must distinguish:

- claimed capacity;
- configured test range;
- bytes written;
- bytes synchronized;
- bytes verified;
- complete and incomplete intervals.

Distributed sampling is useful for diagnosis but cannot certify full capacity.

## 10. Expert Configuration

The TUI exposes complete expert control while also providing coherent presets.

### 10.1 Presets

- **Quick sample:** distributed partial ranges; partial verdict only.
- **Full capacity:** complete write and complete verification.
- **Thermal-safe:** conservative batches with adaptive regulation.
- **Maximum throughput:** increased concurrency and minimal fixed cooldowns.
- **Custom:** fully user-defined.

Selecting a preset populates visible parameters. It does not conceal their values.

### 10.2 Configurable Parameters

- logical pattern block size;
- operating-system I/O request size;
- batch size;
- worker count;
- queue depth;
- exact range or volume to test;
- pattern algorithm and seed;
- cache and synchronization policy;
- metric sampling interval;
- retry count and backoff;
- behavior after an error;
- phase ordering;
- file retention and cleanup behavior;
- fixed cooldown duration;
- thermal thresholds and timing.

The engine validates the complete combination before execution. Unsupported or invalid combinations must be rejected or visibly normalized, never silently ignored.

## 11. Cache and Synchronization Policies

Required policies:

- `buffered`: filesystem-oriented measurement where applicable;
- `sync-per-batch`: recommended default;
- `direct-io`: when correctly supported by platform and target;
- `sync-per-block`: extreme diagnostic mode with an explicit performance warning.

Reports record both the requested policy and the policy actually applied. If an operating system, filesystem, or device cannot honor a requested feature, Nuther marks it unavailable rather than claiming success.

Metrics must distinguish bytes accepted by the application from bytes confirmed after the configured synchronization boundary.

## 12. Thermal Regulation

Thermal behavior has three configurable thresholds:

- **Soft limit:** progressively lower worker count and queue depth.
- **Pause limit:** finish the current safe unit, synchronize, and suspend I/O.
- **Abort limit:** safely terminate when temperature is dangerous or does not recover.

Resume occurs below a separate lower threshold to provide hysteresis and avoid rapid pause/resume oscillation.

Illustrative values:

```text
Soft:   50 °C
Pause:  55 °C
Resume: 48 °C
Abort:  65 °C
```

Defaults may differ for HDD, SATA SSD, and NVMe, but must remain conservative and must not pretend to be vendor-defined limits when authoritative limits are unavailable.

If temperature telemetry is unavailable:

- display `THERMAL TELEMETRY UNAVAILABLE` throughout the run;
- disable adaptive regulation;
- continue with fixed batches and pauses;
- record the degraded protection in all reports;
- never label the run as adaptively thermally protected.

## 13. Checkpoints and Safe Resume

At every completed batch:

1. finish and synchronize the batch;
2. atomically persist the checkpoint;
3. record completed ranges, expected pattern identity, metrics cursor, target identity, and immutable configuration;
4. continue only after checkpoint persistence succeeds.

After interruption:

1. load and validate the checkpoint;
2. re-resolve the stable target identity;
3. reject resume if target or immutable configuration differs;
4. reread and fully verify the last recorded batch;
5. resume at the next batch only after successful verification.

The report preserves separate timeline segments before and after interruption. It must not create an artificial continuous graph.

Checkpoint format must include a schema version and corruption detection. An invalid or ambiguous checkpoint blocks resume and offers a new run instead.

## 14. Temporal Metrics

The default sample interval is two seconds and is configurable in expert mode.

Each sample should include, when applicable:

- monotonic and wall-clock timestamps;
- run phase and regulator state;
- logical offset;
- processed, synchronized, and verified bytes;
- application throughput;
- synchronized write throughput;
- verified read throughput;
- device temperature;
- worker count and queue depth;
- retry count;
- pause reason;
- typed error event.

Statistics are calculated separately for writes, reads, individual batches, and the complete run:

- minimum and maximum;
- arithmetic mean;
- median;
- P5, P50, P95, and P99;
- sustained throughput after excluding explicitly defined warm-up intervals;
- slowest time window;
- variability/stability;
- active duration versus paused duration;
- thermal-performance correlation where sufficient data exists.

The report must document sampling interval, excluded intervals, and calculation rules so numbers remain interpretable.

## 15. Capacity and Integrity Findings

Required findings include:

- claimed target capacity;
- configured and validated capacity;
- first failing offset;
- corrupted blocks and coalesced ranges;
- expected versus observed pattern metadata;
- repeated content from another offset;
- address wraparound evidence;
- short reads and short writes;
- timeouts;
- operating-system and device I/O errors;
- exceptionally slow ranges.

The pattern design must make it computationally practical to regenerate expected data for arbitrary offsets without retaining the entire written dataset.

## 16. TUI Integration

Add a **Capacity Test** tab for the currently selected drive.

### 16.1 Setup Flow

1. Select `Safe file test` or `Destructive raw test`.
2. Select a preset.
3. Review and modify expert parameters.
4. Run preflight.
5. Review capabilities, warnings, duration range, write volume, and SSD endurance impact.
6. Complete the required confirmation.
7. Start the run.

### 16.2 Active Run View

Show:

- current phase and overall progress;
- current batch and logical offset;
- instantaneous, sustained, and synchronized throughput;
- read verification throughput;
- temperature and thermal-regulator state;
- ETA as a range;
- retry and integrity counters;
- recent typed errors;
- `Pause`, `Resume`, and `Abort safely` actions.

### 16.3 Report View

The report view includes:

- capacity and integrity verdict;
- write, read, and temperature curves;
- state timeline for write, cooldown, verify, interruption, and resume;
- aggregate and per-batch statistics;
- slow or failing ranges;
- safety and capability degradations;
- path to the JSON report.

No HTML, CSV, or PNG export is required in V1.

## 17. CLI

The CLI must use exactly the same configuration validation, engine, checkpoint, and report models as the TUI.

Illustrative commands:

```bash
nuther capacity-test --device /dev/sdb --mode raw --profile full
nuther capacity-test --path /mnt/usb --mode file --profile thermal-safe
nuther capacity-test resume <run-id>
nuther capacity-test report <run-id> --json
```

Final flag names may follow the existing command conventions, but semantics must remain consistent across platforms.

## 18. Error Handling

Errors are typed and persisted. Categories include:

- missing prerequisite or privilege;
- invalid or unsupported configuration;
- occupied target or dependent volume;
- active system disk;
- changed target identity;
- insufficient free space;
- read, write, or synchronization failure;
- data corruption or wrong-offset data;
- excessive temperature;
- unavailable device or disconnection;
- user interruption;
- invalid checkpoint;
- cleanup failure.

On abort, the engine attempts to stop at a safe boundary, synchronize where possible, close resources, and persist the current state. If hardware failure prevents this, the report explicitly marks finalization as incomplete.

## 19. Report Persistence and JSON Compatibility

Reports and checkpoints live in a Nuther-owned configuration/data directory, not beside arbitrary user files except for the file-mode run manifest needed to identify its artifacts.

The JSON report must:

- include a schema version;
- separate immutable configuration, discovered capabilities, timeline samples, findings, and final verdict;
- retain requested versus applied settings;
- preserve interruptions and resumed segments;
- support future migration without changing the meaning of old reports;
- avoid secrets and unnecessary personally identifiable host data.

Large timelines may use a streaming writer during execution. Finalization must not require holding the complete report in memory.

## 20. Testing Strategy

### 20.1 Automated Unit and Property Tests

- deterministic generation at arbitrary offsets;
- corruption detection;
- duplicated-range and address-wraparound detection;
- partial reads and writes;
- state-machine transitions;
- checkpoint atomicity and last-batch revalidation;
- statistics and slow-window calculations;
- thermal throttling, pause, hysteresis, and abort behavior using a simulated sensor;
- safety policy decisions;
- JSON round trips and schema migration.

### 20.2 Backend Integration Tests

- file backend against temporary directories and constrained filesystems;
- raw backend against sparse images and virtual devices;
- cache/synchronization capability reporting;
- alignment and range enforcement;
- interruption and resume during both write and verify phases.

### 20.3 Platform Validation

- Linux: loop devices.
- macOS: attached disk images.
- Windows: temporary VHD/VHDX devices.
- Dedicated physical HDD, SATA SSD, NVMe, USB flash, and counterfeit/fault-injection test media where available.

No automated test may access a real raw device unless all of the following are present:

- explicit environment opt-in;
- stable-identity allowlist;
- test-device designation outside CI defaults;
- an independent confirmation mechanism.

## 21. Delivery Sequence

1. Define versioned run, checkpoint, metrics, and report models.
2. Implement deterministic pattern generation and verification.
3. Implement the file backend and full safe-mode engine.
4. Add temporal metrics, statistics, JSON streaming, and TUI curves.
5. Add atomic checkpoints and safe resume.
6. Add thermal sensing and regulation.
7. Implement and validate the Linux raw backend.
8. Implement and validate the macOS raw backend.
9. Implement and validate the Windows raw backend.
10. Harden cross-platform safety checks and conduct physical-device validation.

All three platforms are part of the V1 acceptance target, but the implementation is sequenced so that platform-specific raw I/O is not debugged simultaneously.

## 22. Acceptance Criteria

The feature is ready when:

- a safe file-mode run can fill, synchronize, verify, resume, clean up, and report a configured volume on all three platforms;
- raw runs operate against platform-native virtual devices on Linux, macOS, and Windows;
- the engine detects injected corruption and simulated fake-capacity wraparound;
- the report shows temporal write, synchronized-write, read, and temperature data;
- a run interrupted after a checkpoint revalidates the last batch before resuming;
- fixed cooldown remains active when temperature telemetry is unavailable;
- the active system disk cannot be destructively tested;
- a stable-identity mismatch prevents destructive writes;
- partial tests never receive a full-capacity `PASS`;
- JSON reports are versioned, parseable, and retain all applied settings and degradations;
- destructive physical-device validation is performed on dedicated test hardware under the explicit safety procedure.

## 23. Key Decisions

- Native Go engine, not wrappers around F3/H2testw.
- Common orchestration core with file and raw-device backends.
- Linux, macOS, and Windows are all V1 targets.
- Full expert controls are exposed, supplemented by presets.
- Safe resume revalidates the last completed batch.
- TUI and raw JSON are the only required V1 report outputs.
- Adaptive thermal protection falls back to fixed batch pauses when telemetry is unavailable.
- Destructive testing requires stable-identifier confirmation and refuses the active system disk.
- Full-capacity claims require complete write and verification coverage.
