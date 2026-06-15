## Context

The current command tab can launch interactive Impacket terminals, but automation requires non-interactive execution so output can be captured and stored. This design keeps execution explicit and single-target for the first version.

## Goals / Non-Goals

**Goals:**

- Run operator-selected commands through backend-spawned Impacket processes.
- Capture stdout/stderr and persist run history.
- Avoid exposing secrets in stored command previews.
- Add guardrails around noisy methods.

**Non-Goals:**

- Multi-target orchestration.
- Live streaming output.
- Uploading tools or long-running agents.

## Decisions

1. Start with single-target runs.

   This limits accidental blast radius and keeps UI/run history manageable.

2. Use `exec.CommandContext`.

   The backend can apply a timeout, capture combined output, and store the result. The browser never runs shell commands.

3. Store method and command, not plaintext credential secrets.

   Run history keeps target/credential IDs and redacted command preview. It does not persist passwords beyond existing credential storage.

4. Keep presets read-only.

   Presets are safe triage commands. Custom commands remain possible but require explicit typing and clicking Run.

## Risks / Trade-offs

- Impacket may not be installed -> Backend returns a clear failure and stores the failed run.
- `psexec` is noisy -> UI labels it as noisy and requires explicit method selection.
- Commands can be destructive -> Initial presets are read-only; custom command remains operator-controlled.

## Migration Plan

1. Add remote task run table during database initialization.
2. Add run models, database helpers, and runner logic.
3. Add API endpoints for run and history.
4. Add UI tab and validation.
