## Context

Targets are already stored per team in SQLite and exposed through team-scoped APIs. Network status should reuse those targets, scan only the predefined Windows ports, and present a concise "what is open now" view.

## Goals / Non-Goals

**Goals:**

- Store and display only open tracked ports.
- Scan through backend-owned nmap calls.
- Support manual scans and global interval polling.
- Avoid overlapping background scans.

**Non-Goals:**

- Closed/filtered state history.
- Arbitrary scan profiles.
- Per-team polling intervals.

## Decisions

1. Use `nmap -p 135,139,389,445,3389,5985 --open -oX -`.

   XML output is stable enough to parse without scraping human text. `--open` keeps results focused on actionable ports.

2. Replace findings for scanned targets on each successful scan.

   The database stores currently open ports only. If a port disappears from scan output, it is removed from the visible findings for that target.

3. Use a global polling config.

   The first version keeps one enabled flag and one interval in seconds. This is simpler than per-team schedules and matches the requested configurable interval.

4. Keep polling best-effort.

   If nmap is missing or a scan fails, the backend records scan status metadata and the UI shows the error. Existing findings are not invented or marked closed.

## Risks / Trade-offs

- nmap may not be installed -> Manual scans and polling report a clear error.
- Long scans could overlap -> Background poller skips ticks while a scan is running.
- Scan output can vary by nmap version -> Tests cover the XML fields this feature reads.

## Migration Plan

1. Add open-port and config tables during database initialization.
2. Add scanner parser and scan orchestration.
3. Add APIs for status, manual scan, and config.
4. Add UI tab and controls.
5. Verify tests and frontend check.
