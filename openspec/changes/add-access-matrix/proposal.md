## Why

Operators need a fast way to see which credentials are likely useful against which targets. The app already tracks targets, credentials, Kerberos caches, and open ports, so it can combine that data into an inferred access matrix.

## What Changes

- Add an Access Matrix tab scoped by team.
- Load existing targets, credentials, Kerberos caches, and network status for the selected team.
- Infer likely access methods from credential type, open ports, target metadata, and available Kerberos caches.
- Show compact method badges such as `SMB`, `Dump`, `Exec`, `WinRM`, `RDP`, `LDAP`, `DCOM`, and `Kerberos`.
- Add filters for method, credential type, machine accounts, and rows with possible access.

## Non-goals

- Actively testing whether credentials work.
- Persisting inferred access rows.
- Adding new backend APIs for the first version.
- Replacing the command builder.

## Capabilities

### New Capabilities

- `access-matrix`: Team-scoped inferred credential-to-target access visibility.

### Modified Capabilities

- None.

## Impact

- Frontend adds an Access Matrix tab and client-side inference.
- Existing backend APIs are reused.
- Validation should ensure the Svelte app type-checks.
