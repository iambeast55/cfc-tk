## Why

Operators need a quick view of which known targets currently expose useful Windows services. The app already stores team-scoped targets, so it can periodically scan those targets and show only open, actionable ports.

## What Changes

- Add a Network Status tab showing open tracked ports per target.
- Track only standard Windows ports: `135`, `139`, `389`, `445`, `3389`, and `5985`.
- Add backend nmap scanning that stores only currently open tracked ports.
- Add a configurable global polling interval with enable/disable controls.
- Add manual scan support for the selected team.

## Non-goals

- Full port scanning or arbitrary port lists.
- Displaying closed, filtered, or unknown port states.
- Per-team polling schedules.
- Replacing nmap with a custom scanner.

## Capabilities

### New Capabilities

- `network-status`: Team-scoped visibility into currently open tracked Windows ports.

### Modified Capabilities

- None.

## Impact

- Backend SQLite schema adds open-port findings and network polling config.
- Backend gains nmap execution, XML parsing, polling, and manual scan APIs.
- Svelte UI gains a Network Status tab with team selection, interval controls, and open-port display.
- Tests cover parsing and persistence behavior without requiring nmap to be installed.
