# CFC-ImpUI

CFC-ImpUI is a UI driven tool for Impacket based commands, specifically built for cyber competitions. The idea being to give a way to handle multiple teams at once, and hopefully in a more organized single point manner.

## Features

- Separate team workspaces
- Track domains and targets
- Store credentials and Kerberos cache
- Run `secretsdump` and import the results into the credential table
- Generate Kerberos caches with `getTGT` and `ticketer`
- Launch interactive shells with `wmiexec`, `smbexec`, and `dcomexec`
- Run quick remote tasks with `wmiexec` or `psexec`
- Save notes per team
- Do simple network status checks and port scans

## Requirements

At a minimum:

- Node.js 20+
- npm
- Go 1.24+
- Python 3
- Impacket installed and available in `PATH`
- `libnss-wrapper`

## Installation

Clone the repo, then install the UI and server dependencies.

```sh
npm install
cd server
go mod download
cd ..
```

## Running it

### Option 1:

From the project root:

```sh
./ccfc-impui
```

- install Node packages if needed
- build the UI
- build the Go server
- start the API at `http://localhost:8765`
- serve the UI at `http://127.0.0.1:5173`

### Option 2: dev mode

From the project root:

```sh
npm run dev
```

That starts:

- UI: `http://localhost:5173`
- API: `http://localhost:8765`

### Option 3: run each side separately

UI:

```sh
npm run dev:ui
```

Server:

```sh
npm run server:dev
```

## Build

```sh
npm run build
npm run server:build
```

## Data storage

Local data is stored in:

```text
server/teams.db
```

The database is created automatically if it does not exist.

Kerberos cache files and generated artifacts stay on disk wherever you choose to save them from the UI.

## Project layout

```text
cfc-impui/
|-- src/       # Svelte UI
|-- static/    # Static assets
|-- server/    # Go API, SQLite logic, command runners
|-- scripts/   # Dev helpers
`-- ccfc-impui    # Simple launcher script
```

## Screenshots

Add your screenshots here.

### Dashboard

![Dashboard screenshot](./docs/screenshots/dashboard.png)

### Command tab

![Command tab screenshot](./docs/screenshots/command-tab.png)

### Credentials

![Credentials screenshot](./docs/screenshots/credentials.png)

### Easy mode

![Easy mode screenshot](./docs/screenshots/easy-mode.png)

If you do not want broken image links before adding screenshots, just create the folders and drop files in later.
