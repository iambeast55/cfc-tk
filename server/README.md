# CFC-TK Server

Local Go REST API for CFC-TK team data. It stores teams in a SQLite database.

## Run

From the repository root:

```sh
npm run server:dev
```

From this folder:

```sh
go run .
```

The server starts at `http://localhost:8080`.

## Endpoints

```text
GET    /health
GET    /api/teams
POST   /api/teams
GET    /api/teams/{name}
PUT    /api/teams/{name}
DELETE /api/teams/{name}
```

## Data

The SQLite database lives at `server/teams.db`. It is ignored by git and is created automatically on first run.
