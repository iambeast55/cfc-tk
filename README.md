# CFC-TK

CFC-TK now contains both the Svelte UI and the local Go API server in one project.

## Requirements

- Node.js and npm
- Go 1.24 or newer

## Install

```sh
npm install
cd server
go mod download
```

## Run Everything

From the project root:

```sh
npm run dev
```

This starts:

- UI: `http://localhost:5173`
- API server: `http://localhost:8080`

The UI calls the API at `http://localhost:8080`.

## Run One Side

```sh
npm run dev:ui
npm run server:dev
```

## Build

```sh
npm run build
npm run server:build
```

## Project Layout

```text
cfc-tk/
|-- src/       # Svelte UI
|-- static/    # Static UI assets
|-- server/    # Go REST API and SQLite storage
`-- scripts/   # Development helpers
```

The server stores local data in `server/teams.db`. That file is ignored by git and is created automatically if it does not exist.
