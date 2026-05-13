## Context

The app already treats teams as the primary workspace boundary. Backend data for credentials, domains, targets, and Kerberos caches is stored in SQLite tables keyed by `team_name`, exposed through `/api/teams/{name}/...` routes, and removed with the team through foreign-key cascades. The current Notes tab is only an unbound textarea, so it cannot load different content per team or persist operator notes.

## Goals / Non-Goals

**Goals:**

- Persist one plain-text notes document per team in the existing SQLite database.
- Expose team notes through the existing team-scoped API route pattern.
- Keep the Notes tab fast and predictable: select team, edit notes, save, and see errors or success state.
- Preserve team isolation and delete notes automatically when a team is deleted.

**Non-Goals:**

- Rich-text editing, multiple note documents, attachments, search, or note history.
- Real-time collaboration or concurrent edit conflict resolution.
- External storage or a separate notes service.

## Decisions

1. Store notes in a dedicated `team_notes` table.

   Use `team_name TEXT PRIMARY KEY`, `content TEXT NOT NULL DEFAULT ''`, `updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP`, and a foreign key to `teams(name) ON DELETE CASCADE`. This keeps notes independent from the legacy `teams.targets` JSON column and avoids overloading the `teams` row with a potentially large text field.

   Alternative considered: add a `notes` column directly to `teams`. That would be simpler, but a separate table mirrors the existing team-scoped resource model and keeps future metadata such as `updated_at` cleanly scoped.

2. Use idempotent upsert semantics for saves.

   A `PUT /api/teams/{name}/notes` endpoint will create or replace the team note content after verifying the team exists. This lets the frontend save the textarea without needing to know whether the note row already exists.

   Alternative considered: `POST` to create and `PATCH` to update. That adds statefulness that is unnecessary for a single note document per team.

3. Return an empty note for existing teams with no saved note row.

   `GET /api/teams/{name}/notes` will return `{ "teamName": "...", "content": "", "updatedAt": "" }` or equivalent for teams that exist but do not yet have saved notes. Missing teams continue to return `404`.

   Alternative considered: return `404` for missing notes. That would make a new team's first Notes load look like an error even though the team is valid.

4. Keep the frontend save explicit.

   The Notes tab should load notes when the selected team changes and save when the operator clicks a save button. This avoids hidden writes while typing and matches field workflow where a user may draft before committing.

   Alternative considered: autosave on debounce. That is convenient but introduces harder-to-see failure states and potential race behavior without materially improving the initial capability.

## Risks / Trade-offs

- Lost edits from two browser windows saving the same team's notes -> Mitigation: initial implementation uses last-write-wins and clear save status; conflict detection is out of scope.
- Large note bodies could make the Notes tab sluggish -> Mitigation: store as SQLite `TEXT` and keep the editor plain text; no rendering pipeline is introduced.
- Existing databases need schema update -> Mitigation: create `team_notes` with `CREATE TABLE IF NOT EXISTS` during `initDB`, consistent with existing tables.
- Team names are path parameters -> Mitigation: continue using existing encoded team-name route patterns and server-side team lookup before reading or writing notes.

## Migration Plan

1. Add the `team_notes` table during database initialization.
2. Add model, request/response, and database helpers for reading and saving team notes.
3. Add handlers and router entries for `GET` and `PUT /api/teams/{name}/notes`.
4. Update the Notes tab to select a team, load notes, edit content, save content, and show busy/error/saved states.
5. Verify existing teams with no note row show empty notes, and deleting a team removes its note row through cascade.

Rollback is to remove the route/UI usage and leave the unused table in place; no destructive migration is required.

## Open Questions

- None for the initial plain-text, single-note-per-team implementation.
