## Why

The Notes tab currently provides a scratch textarea that is not tied to a team and does not survive page reloads. Red team operators need persistent, team-specific notes so observations, reminders, and triage details stay attached to the correct blue team workspace.

## What Changes

- Add persistent notes per team using the existing SQLite database.
- Add backend API endpoints under the existing team route shape to read and save a team's note content.
- Update the Notes tab to let the operator select a team, load its saved notes, edit them, and save changes.
- Keep notes isolated by team and remove them automatically when the owning team is deleted.
- Preserve the current single-note editing experience; this change does not introduce multi-note notebooks or rich-text editing.

## Non-goals

- Adding collaborative editing, note history, or conflict resolution.
- Adding rich text, attachments, tagging, or search.
- Moving notes to a separate database or external storage service.
- Changing existing credential, target, domain, or Kerberos cache behavior.

## Capabilities

### New Capabilities

- `team-notes`: Persistent plain-text notes scoped to individual teams.

### Modified Capabilities

- None.

## Impact

- Backend SQLite schema gains a team-notes table or equivalent team-scoped storage with a foreign key to `teams(name)`.
- Backend models, database helpers, handlers, and router gain team notes read/save support.
- Svelte Notes tab gains team selection, load/save state, errors, and persisted textarea binding.
- Tests should cover database persistence, team isolation, missing-team handling, and frontend save/load behavior where practical.
