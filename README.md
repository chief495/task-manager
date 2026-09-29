# Task Manager

Task Manager is a Go and TypeScript web application for project work tracking, tasks, reports, contacts, project stages, labor calculations, and timesheets.

## Tech Stack

- Go 1.25
- Gin
- GORM
- PostgreSQL
- TypeScript

## Features

- User authentication and role-based navigation
- Project, phase, task, and report management
- Workflow stages and statuses with transitions
- Organization and contact tracking
- Labor norms and calculation tools
- Timesheet and export workflows

## Getting Started

1. Copy environment defaults:

```sh
cp .env.example .env
```

2. Start PostgreSQL:

```sh
docker compose up -d
```

3. Install frontend tooling:

```sh
npm install
```

4. Build frontend assets:

```sh
npm run build
```

5. Run the app:

```sh
make run-debug
```

The application listens on `http://localhost:8080` by default.

## Development

```sh
make build
make test
make deps-check
```

## Repository Hygiene

Generated binaries, local environment files, IDE settings, certificates, and Go build cache are intentionally ignored.

## License

MIT
