# CFCBench

CFCBench is a UI driven tool for Impacket based commands, specifically built for cyber competitions. The idea being to give a way to handle multiple teams at once, and hopefully in a more organized single point manner.

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
- Open SSH terminals for Linux targets with password or private-key authentication
- Upload and download individual files with SCP
- Stop background network polling immediately from Network

## Requirements

At a minimum:

- Node.js 22+
- npm
- Go 1.24+
- Python 3
- Impacket installed and available in `PATH`
- `libnss-wrapper`

## Installation

For a Linux deployment archive, run `python scripts/package_linux.py`.
The archive and SHA-256 checksum are written to `dist/`. Extract the archive on
Linux, then run `./install.sh` and `./run.sh`. See `deploy/linux/README.md` for
dependencies and deployment details. Local databases and caches are excluded.

Clone the repo, then install the UI and server dependencies.

```sh
git clone https://github.com/iambeast55/cfc-tk.git cfcbench
cd cfcbench
./cfcbench
```

The launcher installs Node dependencies, builds both components, and starts
the app. For SSH/SCP on Kali, install `openssh-client` and `sshpass`; interactive
terminals require starting CFCBench inside your graphical desktop session.
Network scans require `nmap`.

To install dependencies separately:

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
./cfcbench
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

### Config import

On Main, use Import config to upload a JSON file (up to 2 MB). Download the
example there, or start from `static/config-example.json`. The format uses
`version: 1` and a `teams` array; each team has `name`, `subnetId`, optional
`domains` (names), and `targets` with `hostname`, `ip`, `os` (windows/linux),
and optional `domain`. Omit domain for standalone hosts. Domain names referenced
by targets are created automatically. Exact existing records are skipped;
conflicts or invalid entries cancel the entire import. Credentials and command
execution settings are not imported.

Local data is stored in:

```text
server/teams.db
```

The database is created automatically if it does not exist.

Kerberos cache files and generated artifacts stay on disk wherever you choose to save them from the UI.

## Project layout

```text
cfcbench/
|-- src/       # Svelte UI
|-- static/    # Static assets
|-- server/    # Go API, SQLite logic, command runners
|-- scripts/   # Dev helpers
`-- cfcbench    # Simple launcher script
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
