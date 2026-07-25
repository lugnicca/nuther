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

## 5.1 Normative Safety Invariants

The following rules are non-overridable, including in expert mode:

- Destructive mode fails closed whenever target identity, storage topology, active use, exclusive access, size, alignment, or range bounds cannot be established conclusively.
- Safety checks are repeated immediately before the first write, after every reopen or reconnection, and before any resumed write.
- A hot-plug event, identity ambiguity, topology change, failed required flush, checkpoint failure, thermal abort, or range violation terminates writes without automatic continuation.
- Destructive byte ranges are overflow-checked half-open intervals `[start, end)` relative to one precisely identified target object. They are never rounded outward.
- Expert settings may reduce performance or diagnostic coverage, but may not weaken identity checks, authorized bounds, system-disk protection, required synchronization, checkpoint integrity, or thermal hard stops.
- Unsupported combinations are rejected during preflight. Nuther never silently normalizes a destructive range or safety-relevant option.

## 5.2 Platform Capability Contract

Each platform adapter must publish a capability record covering:

- supported raw target objects: physical disk, partition, or platform volume;
- stable identity sources and confidence level;
- storage-topology and physical-ancestor discovery;
- exclusive locking/offline behavior;
- sector geometry and alignment;
- buffered, direct, flush, and cache-control guarantees;
- filesystem allocation reporting;
- temperature sensors and freshness;
- supported safety degradations and their verdict consequences.

Equivalent capability states must lead to equivalent preflight and verdict decisions across operating systems. A release-owned matrix documents the concrete Linux, macOS, and Windows APIs and the combinations eligible for certifying verdicts.

## 6. Run Lifecycle

Every run is a persisted state machine.

### 6.1 Preflight

Before any write, Nuther must:

- resolve a stable device identity;
- record model, serial, capacity, transport, logical sector size, and physical sector size when available;
- resolve all active system resources—including root/system, boot/EFI, recovery when active, executable and data volumes used by Nuther, swap/pagefile, and hibernation—to every physical storage ancestor;
- reject raw mode if the selected target is any such ancestor, or if ancestry cannot be determined conclusively;
- determine privileges and backend capabilities;
- detect mounted partitions, open volumes, swap/paging use, encryption layers, RAID/LVM/device-mapper/APFS/Storage Spaces membership, pools, virtual-disk backing, and dependent volumes; inconclusive discovery rejects raw mode;
- calculate the exact target range and expected write volume;
- validate alignment, block size, concurrency, cache, and synchronization settings;
- probe temperature telemetry availability;
- estimate duration as a range rather than a falsely precise timestamp;
- estimate device write impact, especially for SSDs;
- acquire exclusive access for raw mode and pin the opened object to the resolved identity;
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

Certifying verification requires the write phase to finish its required synchronization boundary, close and reopen the target, revalidate identity/topology/size, and use the strongest platform-supported cache-bypass or cache-invalidation mechanism. If Nuther cannot establish that reads are not satisfied solely by the application page cache, the report records `verification_strength: cached_or_unknown` and the run cannot receive a capacity-certifying `PASS`.

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

Verdicts have deterministic precedence: `SAFETY_ABORT` > `DATA_CORRUPTION` or `CAPACITY_MISMATCH` > `IO_FAILURE` > `THERMAL_ABORT` > `USER_ABORT` > `INCOMPLETE` > `PARTIAL_PASS` > `PASS`. Integrity verdict, run completion status, finalization status, and cleanup status are also stored as separate fields so cleanup failure cannot hide a successful or failed integrity result. `PARTIAL_PASS` means every byte in the configured partial range met the required synchronization and verification strength, while less than the full certifiable target was covered.

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

- On the active system volume, enforce a minimum free-space reserve that cannot be disabled: `max(5% of filesystem capacity, 10 GiB)` plus a platform/filesystem metadata allowance. Other volumes default to `max(5%, 1 GiB)`.
- Refuse configurations that would cross the enforced reserve.
- Requery filesystem-local available space before every allocation unit and batch; stop before crossing the reserve.
- Treat inability to obtain reliable availability as a preflight failure on the active system volume.
- Treat `ENOSPC` or its platform equivalent as an immediate safety stop, never as a retryable write error.
- Use a recognizable Nuther directory and manifest so stale files can be attributed and safely recovered.

The file backend must securely create a new, unpredictable run directory without following symlinks, junctions, reparse points, or mount-point substitutions. It pins filesystem/volume identity and uses handle-relative operations where available. Every artifact is recorded in the manifest with identity and ownership metadata. Cleanup deletes only artifacts whose run identity, manifest entry, ownership, and pinned filesystem location all match; otherwise cleanup is refused and reported.

Writes must be non-sparse. The report records logical bytes and allocated bytes separately and detects known compression, deduplication, quotas, copy-on-write, and thin-provisioning conditions where APIs permit. A certifying file-mode verdict requires verified allocated extents. If physical allocation cannot be established, the run receives a degraded, non-certifying verification result rather than a full capacity `PASS`.

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

Stable identity is a versioned tuple with a confidence classification. It combines the strongest available immutable hardware identifier, capacity, logical and physical sector geometry, transport identity, and bus/location information where appropriate. Serial-only identity is insufficient when absent, duplicated, truncated, or bridge-generated. Destructive mode is rejected when the tuple is absent, ambiguous, or non-unique. Identity is revalidated on every open/reopen and after any device event; disconnect/reconnect terminates the run and requires an explicit resume flow.

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

All ranges are overflow-checked half-open byte intervals `[start, end)` relative to a named target object. Reports and destructive confirmations show exact byte bounds and inclusive LBA coverage. Start and end must already satisfy sector and backend alignment; Nuther rejects rather than rounds outward. Size and bounds are freshly queried immediately before writing and after every reopen.

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
- approved certification pattern and seed;
- cache and synchronization policy;
- metric sampling interval;
- retry count and backoff;
- bounded retry policy for explicitly retryable short/transient I/O errors;
- file retention and cleanup behavior;
- fixed cooldown duration;
- thermal thresholds and timing.

V1 supports only these phase sequences: `write -> verify`, `write-only` with a non-certifying result, and `verify-resume` for data written by the same compatible run. Arbitrary phase ordering and user-defined continuation after errors are out of scope. The engine validates the complete combination before execution. Unsupported or invalid combinations are rejected. Non-safety performance values may be visibly clamped inward only when the resulting value remains inside already confirmed bounds.

Non-overridable fatal categories are safety uncertainty, identity or topology change, unauthorized range, thermal abort, checkpoint commit failure, required synchronization failure, and wrong-offset data. Retries are allowed only for classified transient operations, must preserve the exact offset and remaining byte count, and are bounded. Diagnostic reading may continue after corruption when explicitly selected, but cannot restore `PASS` and cannot authorize additional writes.

## 11. Cache and Synchronization Policies

Required policies:

- `buffered`: filesystem-oriented measurement where applicable;
- `sync-per-batch`: recommended default;
- `direct-io`: when correctly supported by platform and target;
- `sync-per-block`: extreme diagnostic mode with an explicit performance warning.

Reports record both the requested policy and the policy actually applied. If an operating system, filesystem, or device cannot honor a requested feature, Nuther marks it unavailable rather than claiming success.

Metrics must distinguish bytes accepted by the application from bytes confirmed after the configured synchronization boundary.

The report records distinct durability levels: `application_accepted`, `os_flush_acknowledged`, `device_flush_requested`, `device_flush_acknowledged`, and `unsupported_or_unknown`. Platform adapters document what each API can actually prove. A failed or unsupported boundary required by the selected certifying profile prevents `PASS`; Nuther must not label data physically durable when the device or bridge cannot acknowledge that guarantee.

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

The fallback is mandatory and non-zero. Initial conservative defaults are batches no larger than 32 GiB followed by at least 60 seconds of pause for HDD/SATA SSD, and batches no larger than 16 GiB followed by at least 90 seconds for NVMe; these defaults are subject to validation on real hardware before release. When telemetry is unavailable or becomes invalid, continuous mode and values below the validated fallback minima are rejected or suspended until the fallback is applied. The maximum-throughput preset visibly degrades to this fallback.

The thermal adapter reports all relevant drive/controller sensors, their source, update time, and validity. Regulation uses the hottest valid relevant sensor. Readings outside documented bounds, repeated errors, or readings older than three expected sampling periods are stale. Sensor loss or staleness during a run immediately transitions to the fixed fallback; if the fallback cannot be safely entered, the run aborts. The transition is persisted in timeline and report degradations.

## 13. Checkpoints and Safe Resume

At every completed batch:

1. finish and synchronize the batch;
2. atomically persist the checkpoint;
3. record completed ranges, expected pattern identity, metrics cursor, target identity, and immutable configuration;
4. continue only after checkpoint persistence succeeds.

The checkpoint stores separate written, synchronization-confirmed, and verified interval sets/frontiers; active phase; partial unit; timeline generation; and finalization state. The restart contract explicitly covers interruption before write, during partial write, after synchronization but before checkpoint publication, during verification, and during finalization. Nuther revalidates the last committed safe unit of the interrupted phase and continues from that phase's frontier.

Checkpoint commit uses generation-numbered files: write the next generation, flush contents, verify checksum, atomically publish with the platform-supported replacement primitive, then durably commit the containing directory or documented platform equivalent. Checkpoint and streamed timeline generations must agree. Recovery may fall back only to the newest complete checksum-valid matching generation; it never reconstructs or guesses missing state.

After interruption:

1. load and validate the checkpoint;
2. re-resolve the stable target identity;
3. for raw mode, rerun the complete current safety preflight, including system ancestry, active dependencies, privileges, exclusive access, capacity, target bounds, and identity confidence;
4. reject resume if target, topology, bounds, safety state, or immutable configuration differs or is inconclusive;
5. require the same destructive authorization contract again, including interactive confirmation when resumed from the TUI;
6. reread and fully verify the last committed safe unit for the interrupted phase;
7. resume from that phase's next safe frontier only after successful verification.

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

Normative calculations use monotonic elapsed time and byte deltas. Instantaneous throughput is bytes completed during one sample interval divided by active monotonic seconds. Aggregate throughput is total bytes divided by active phase time and excludes declared pause intervals. Percentiles operate on fixed-duration time samples and are time-weighted; batches and requests are reported separately rather than mixed into that population. The default sustained metric excludes the first 30 seconds or first 1 GiB of a phase, whichever completes first. The slowest window uses a fixed 60-second active-I/O window. Synchronization latency is reported separately; bytes become synchronization-confirmed only at the successful boundary. Golden-data tests define every formula and remain stable when the display sampling interval changes.

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

V1 includes one mandatory, versioned certification pattern. Each independently verifiable block binds pattern version, run ID, stable target identity digest, absolute offset, payload length, and seed, and contains an offset-addressable pseudorandom payload plus a strong payload checksum. The specification and test vectors for that pattern are part of implementation planning. Only approved certification patterns may produce a certifying verdict; experimental or compressible patterns are diagnostic-only.

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

Shipping destructive raw mode requires a documented release-gate matrix with at least one real removable device tested on Linux, macOS, and Windows. Each platform must include negative evidence for active-system ancestry, mounted or pooled dependencies, unavailable or ambiguous identity, failed exclusive access, alias paths, identity mismatch, disconnect/reconnect, non-aligned bounds, size change, and inconclusive topology. Any unresolved fail-closed safety case blocks raw-mode release on that platform; file mode may ship independently.

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
- raw testing fails closed when physical ancestry or active-use topology is inconclusive;
- every raw resume repeats the complete safety preflight and destructive authorization;
- identity remains pinned through exclusive access and is revalidated after every reopen;
- exact half-open bounds cannot be rounded or written outside the confirmed range;
- file-mode reserve checks prevent active-volume exhaustion races and path substitution;
- certifying verification cannot be satisfied solely from the application page cache;
- checkpoint and timeline generations recover deterministically after simulated power-loss points;
- mandatory non-zero thermal fallback activates when telemetry is absent, stale, or lost;
- aggregate metric formulas pass fixed golden-data tests;
- a stable-identity mismatch prevents destructive writes;
- partial tests never receive a full-capacity `PASS`;
- JSON reports are versioned, parseable, and retain all applied settings and degradations;
- destructive physical-device validation is performed on dedicated test hardware under the explicit safety procedure.

## 23. V1 Compatibility Boundaries

To keep the expert surface testable, V1 certifying profiles permit only approved combinations from a maintained compatibility matrix. The matrix covers backend, target object, phase sequence, pattern, synchronization level, verification strength, I/O mode, alignment, concurrency range, thermal mode, and retry policy. Arbitrary patterns, arbitrary phase ordering, unbounded retries, and continuation after fatal errors are explicitly excluded. Every permitted combination must map to acceptance coverage; every other combination is a preflight error.

## 24. Key Decisions

- Native Go engine, not wrappers around F3/H2testw.
- Common orchestration core with file and raw-device backends.
- Linux, macOS, and Windows are all V1 targets.
- Full expert controls are exposed, supplemented by presets.
- Safe resume revalidates the last completed batch.
- TUI and raw JSON are the only required V1 report outputs.
- Adaptive thermal protection falls back to fixed batch pauses when telemetry is unavailable.
- Destructive testing requires stable-identifier confirmation and refuses the active system disk.
- Full-capacity claims require complete write and verification coverage.
