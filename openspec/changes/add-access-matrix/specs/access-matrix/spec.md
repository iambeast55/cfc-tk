## ADDED Requirements

### Requirement: Access matrix derives likely access per team
The system SHALL derive likely credential-to-target access methods for the selected team using existing targets, credentials, Kerberos caches, and network status data.

#### Scenario: SMB access inferred
- **WHEN** a target has port `445` open and a credential has password or NTLM material
- **THEN** the matrix shows SMB-related inferred methods for that credential-target pair

#### Scenario: Kerberos access inferred
- **WHEN** a usable Kerberos cache exists for a credential identity
- **THEN** the matrix shows Kerberos as a possible method where target context supports it

### Requirement: Matrix labels inferred confidence
The system SHALL label inferred access methods with confidence levels.

#### Scenario: Matching context
- **WHEN** credential context matches the target domain, host, or IP
- **THEN** the inferred methods are shown with stronger confidence than generic type-and-port matches

### Requirement: Matrix can be filtered
The Access Matrix tab SHALL provide filters for method, credential type, machine accounts, and rows with possible access.

#### Scenario: Filter by method
- **WHEN** an operator selects a method filter
- **THEN** the matrix shows only rows or cells containing that method

### Requirement: Empty states guide operators
The Access Matrix tab SHALL explain when required source data is missing.

#### Scenario: Missing network scan
- **WHEN** no network scan data is available
- **THEN** the tab indicates that scanning Network Status improves suggestions
