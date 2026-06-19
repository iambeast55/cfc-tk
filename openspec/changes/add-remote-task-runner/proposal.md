## Why

Operators need a repeatable way to run explicit one-off commands on known targets using known credentials. The app already stores teams, targets, credentials, and command-building patterns, so it can add a captured remote task runner without opening an interactive terminal.

## What Changes

- Add a Remote Tasks tab for single-target command execution.
- Support `wmiexec` first for captured, non-interactive command output.
- Add `psexec` as an explicit method with a warning label because it is louder.
- Store task run history with method, target, credential, command, status, output, error, and timestamps.
- Add read-only presets such as `whoami`, `hostname`, `ipconfig /all`, `query user`, `tasklist`, and local administrators.
- Require explicit operator action for every run.

## Non-goals

- Batch execution across many targets.
- Automatic command execution from scans or credential imports.
- File upload, persistence, scheduled tasks, or payload staging.
- Interactive shell handling.

## Capabilities

### New Capabilities

- `remote-task-runner`: Explicit remote command execution with captured output and run history.

### Modified Capabilities

- None.

## Impact

- Backend SQLite schema gains remote task run history.
- Backend gains Impacket process execution with timeout and output capture.
- Frontend gains a Remote Tasks tab with team, target, credential, method, command preset, output, and run history.
- Tests cover command construction and history persistence without requiring Impacket to run.
