# gator

A command-line application for managing user authentication and database operations, built with Go and PostgreSQL.

## Features

- User login and registration
- PostgreSQL database integration
- Configuration management
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

3. Set up the database:
```bash
# Create a PostgreSQL database and run migrations
psql -U postgres -d gator < sql/schema/001_user.sql
```

## Usage

### Register a new user
```bash
go run main.go register
```

### Login
```bash
go run main.go login
```

## Building

```bash
go build
```

## Dependencies

- `github.com/google/uuid` - UUID generation
- `github.com/lib/pq` - PostgreSQL driver for Go
