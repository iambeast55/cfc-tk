## 1. Backend Storage

- [x] 1.1 Add SQLite tables for open-port findings, scan metadata, and global polling config.
- [x] 1.2 Add Go models for open-port rows, team network status, scan results, and config.
- [x] 1.3 Add database helpers to replace scanned target findings and read team network status.

## 2. Scanner And API

- [x] 2.1 Add nmap XML parser for open tracked ports.
- [x] 2.2 Add scanner orchestration for manual team scans.
- [x] 2.3 Add background polling using the configured interval.
- [x] 2.4 Add API endpoints for status, scan-now, and config.
- [x] 2.5 Add focused backend tests for parser and persistence behavior.

## 3. Frontend

- [x] 3.1 Add Network Status tab state, load functions, scan action, and config save.
- [x] 3.2 Add Network Status tab UI with team selector, polling controls, scan button, and open-port badges.

## 4. Verification

- [x] 4.1 Run backend tests and frontend validation.
