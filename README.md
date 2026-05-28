# gator

A command-line application for managing user authentication and database operations, built with Go and PostgreSQL.

## Features

- User login and registration
- List all users
- PostgreSQL database integration
- Configuration management via `~/.gatorconfig.json`
- Command-based CLI interface
- Type-safe database queries via sqlc

## Project Structure

```
gator/
├── internal/
│   ├── config/          # Configuration and command handling
│   └── database/        # Database models and generated code
├── sql/
│   ├── schema/          # Database schema definitions
│   └── queries/         # SQL query definitions
├── main.go              # Application entry point
├── sqlc.yaml            # sqlc configuration
└── go.mod               # Go module definition
```

## Requirements

- Go 1.24.4 or later
- PostgreSQL
- sqlc CLI (for code generation)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/peter-njuku/gator.git
cd gator
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the PostgreSQL database:
```bash
createdb gator
psql -d gator < sql/schema/001_user.sql
```

4. Create a config file at `~/.gatorconfig.json` with your database URL:
```json
{
  "db_url": "postgres://user:password@localhost/gator?sslmode=disable",
  "username": ""
}
```

## Usage

The application uses commands in the form `go run . <command> [args]`.

### Register a new user
```bash
go run . register alice
```

### Login as an existing user
```bash
go run . login alice
```

### List all users
```bash
go run . users
```

### Reset the database
```bash
go run . reset
```

## Building

Build the application:
```bash
go build -o gator
```

Then run it:
```bash
./gator users
```

## Dependencies

- `github.com/google/uuid` - UUID generation
- `github.com/lib/pq` - PostgreSQL driver for Go
