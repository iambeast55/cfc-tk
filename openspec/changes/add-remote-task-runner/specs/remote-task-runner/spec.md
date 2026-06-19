## ADDED Requirements

### Requirement: Remote tasks run explicitly
The system SHALL run a remote task only after an operator explicitly selects a team, target, credential, method, and command.

#### Scenario: Run selected command
- **WHEN** an operator submits a remote task run
- **THEN** the backend executes the selected command against the selected target using the selected credential

### Requirement: Remote task output is captured
The system SHALL capture remote task output, error text, status, and timestamps.

#### Scenario: Successful run
- **WHEN** a remote task process exits successfully
- **THEN** the run history records succeeded status and captured output

#### Scenario: Failed run
- **WHEN** a remote task process fails or times out
- **THEN** the run history records failed status and the error/output text

### Requirement: Credentials are not exposed in history
The system SHALL avoid storing plaintext credential secrets in command previews or history rows.

#### Scenario: Password credential
- **WHEN** a password-backed task is run
- **THEN** the stored command preview redacts the password

### Requirement: Remote Tasks tab manages single-target runs
The UI SHALL provide a team-scoped Remote Tasks tab for single-target remote command execution and history review.

#### Scenario: Preset command selected
- **WHEN** an operator selects a command preset
- **THEN** the command input is filled with that preset command
