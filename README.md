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

- **Go 1.24.4** or later
- **PostgreSQL** 12 or later

Ensure both are installed and accessible from your command line before proceeding.

## Installation

### 1. Install the gator CLI

Use `go install` to install the gator CLI directly:

```bash
go install github.com/peter-njuku/gator@latest
```

This will install the `gator` executable in your `$GOPATH/bin` directory (typically `~/go/bin`). Make sure this directory is in your `$PATH`.

### 2. Set up PostgreSQL

Create a new database for gator:

```bash
createdb gator
```

Initialize the database schema:

```bash
psql -d gator < sql/schema/001_user.sql
```

(You'll need to clone the repository to access the schema file if you haven't already.)

### 3. Configure gator

Create a configuration file at `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://user:password@localhost/gator?sslmode=disable",
  "username": ""
}
```

Replace `user` and `password` with your PostgreSQL credentials. The `username` field will be populated when you log in.

## Usage

Once installed, run commands using:

```bash
gator <command> [args]
```

### Available Commands

**Register a new user:**
```bash
gator register alice
```

**Login as an existing user:**
```bash
gator login alice
```

**List all users:**
```bash
gator users
```

**Reset the database:**
```bash
gator reset
```

## Building from Source

If you prefer to build locally instead of using `go install`:

```bash
git clone https://github.com/peter-njuku/gator.git
cd gator
go build -o gator
./gator <command>
```

## Dependencies

- `github.com/google/uuid` - UUID generation
- `github.com/lib/pq` - PostgreSQL driver for Go
