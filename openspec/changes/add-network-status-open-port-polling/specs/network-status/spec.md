## ADDED Requirements

### Requirement: Open tracked ports are stored per team target
The system SHALL store only open findings for tracked Windows ports on known team targets.

#### Scenario: Open ports are saved
- **WHEN** a scan finds tracked ports open on a target
- **THEN** the system persists those open ports for that target

#### Scenario: Missing ports are not displayed
- **WHEN** a scan does not report a tracked port as open
- **THEN** the system does not display that port as open

### Requirement: Network scans use nmap
The backend SHALL run nmap for tracked ports `135`, `139`, `389`, `445`, `3389`, and `5985` and parse structured output.

#### Scenario: Manual scan
- **WHEN** an operator starts a manual network status scan for a team
- **THEN** the backend scans that team's known target IPs and stores currently open tracked ports

#### Scenario: Scanner failure
- **WHEN** nmap is unavailable or fails
- **THEN** the backend returns a scan error without fabricating open ports

### Requirement: Polling is configurable
The system SHALL allow network status polling to be enabled or disabled and SHALL allow the polling interval to be configured.

#### Scenario: Polling enabled
- **WHEN** polling is enabled
- **THEN** the backend periodically scans known targets at the configured interval

#### Scenario: Polling disabled
- **WHEN** polling is disabled
- **THEN** the backend does not run scheduled scans

### Requirement: Network Status tab shows open ports
The UI SHALL show known targets for the selected team and list only their currently open tracked ports.

#### Scenario: Target has open ports
- **WHEN** a selected team target has stored open-port findings
- **THEN** the tab shows those ports as compact open-port badges

#### Scenario: Target has no open tracked ports
- **WHEN** a selected team target has no stored open-port findings
- **THEN** the tab shows that no tracked ports are currently open
