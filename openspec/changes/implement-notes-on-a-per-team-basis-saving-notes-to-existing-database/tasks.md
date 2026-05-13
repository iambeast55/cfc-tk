## 1. Backend Data Model

- [x] 1.1 Add a `team_notes` SQLite table during database initialization with `team_name`, `content`, `updated_at`, and an `ON DELETE CASCADE` foreign key; acceptance: existing databases initialize without errors and team deletion removes note rows.
- [x] 1.2 Add Go request/response models for team notes; acceptance: JSON fields include team name, content, and update metadata using existing model naming conventions.
- [x] 1.3 Add database helpers to get and upsert notes by team name; acceptance: helpers return empty content for existing teams with no note row, save updates idempotently, and return not found for missing teams.

## 2. Backend API

- [x] 2.1 Add `GET /api/teams/{name}/notes` handler and router entry; acceptance: existing teams return saved or empty notes and missing teams return 404.
- [x] 2.2 Add `PUT /api/teams/{name}/notes` handler and router entry; acceptance: valid content is persisted and returned, invalid or missing teams produce appropriate error responses.
- [x] 2.3 Add focused backend tests for note persistence, team isolation, missing-team behavior, and cascade cleanup; acceptance: `go test ./...` passes.

## 3. Frontend Notes Tab

- [x] 3.1 Add Notes tab state for selected team, note content, loading, saving, saved status, dirty state, and errors; acceptance: state resets safely when teams are unavailable or selected team changes.
- [x] 3.2 Load notes from the backend when the Notes tab has a selected team; acceptance: the textarea shows the selected team's persisted notes and displays load failures.
- [x] 3.3 Save edited notes through the backend `PUT` endpoint; acceptance: clicking save persists the content, updates saved status, and reports save failures without losing textarea content.
- [x] 3.4 Update the Notes tab UI with a team selector and disabled empty state; acceptance: operators can switch teams, see each team's note content, and cannot edit/save when no team exists.

## 4. Verification

- [x] 4.1 Run backend tests and frontend validation available in the repo; acceptance: relevant test, check, or build commands pass or any unrelated failures are documented.
- [x] 4.2 Manually verify the notes workflow against a running backend and UI; acceptance: create/select two teams, save different notes for each, reload, and confirm notes remain isolated.
