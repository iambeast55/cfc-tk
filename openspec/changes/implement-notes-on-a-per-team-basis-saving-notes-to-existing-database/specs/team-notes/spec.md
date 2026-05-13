## ADDED Requirements

### Requirement: Team notes are persisted per team
The system SHALL store a plain-text notes document for each team in the existing database and SHALL keep each team's notes isolated from all other teams.

#### Scenario: Saving notes for a team
- **WHEN** an operator saves notes for an existing team
- **THEN** the system persists the submitted note content for that team
- **AND** subsequent loads for that team return the saved content

#### Scenario: Team isolation
- **WHEN** notes are saved for one team
- **THEN** loading notes for a different team returns only that different team's content

#### Scenario: New team without notes
- **WHEN** an operator loads notes for an existing team that has never saved notes
- **THEN** the system returns an empty note document for that team

### Requirement: Team notes use team-scoped API routes
The system SHALL expose team notes through API routes scoped by team name and SHALL reject note operations for teams that do not exist.

#### Scenario: Load team notes
- **WHEN** the frontend requests notes for an existing team
- **THEN** the backend returns the team name, note content, and note update metadata

#### Scenario: Save team notes
- **WHEN** the frontend submits note content for an existing team
- **THEN** the backend creates or updates that team's note record and returns the saved note

#### Scenario: Missing team
- **WHEN** the frontend loads or saves notes for a team name that does not exist
- **THEN** the backend responds with a not found error

### Requirement: Team notes are removed with their team
The system SHALL delete a team's saved notes when the owning team is deleted.

#### Scenario: Delete team with notes
- **WHEN** a team with saved notes is deleted
- **THEN** the team's notes are removed from the database
- **AND** future note requests for that team fail because the team no longer exists

### Requirement: Notes tab edits the selected team's notes
The Notes tab SHALL allow the operator to select a team, load that team's saved notes into a plain-text editor, and explicitly save changes back to the database.

#### Scenario: Select team loads notes
- **WHEN** the operator opens the Notes tab and selects a team
- **THEN** the editor displays the notes saved for that team

#### Scenario: Save edited notes
- **WHEN** the operator edits the note content and clicks save
- **THEN** the frontend sends the updated content to the backend
- **AND** the tab indicates whether the save succeeded or failed

#### Scenario: No teams available
- **WHEN** no teams exist
- **THEN** the Notes tab disables editing and indicates that a team must be created before notes can be saved
