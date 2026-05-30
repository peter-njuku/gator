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

## Development

### Setting up the development environment

1. Clone the repository:
```bash
git clone https://github.com/peter-njuku/gator.git
cd gator
```

2. Install dependencies:
```bash
go mod download
```

3. Set up PostgreSQL as described above.

### Working with the database using sqlc

This project uses **sqlc** to generate type-safe Go code from SQL queries. This means:
- SQL queries are written in plain SQL
- sqlc automatically generates Go code with proper types and error handling
- Changes are tracked and versioned alongside your schema

#### Prerequisites for development

Install sqlc:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

#### Adding new database features

**1. Create a new migration** (if modifying the schema):

Add a new file in `sql/schema/` following the naming convention `002_<feature>.sql`:

```sql
-- sql/schema/002_posts.sql
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

Apply the migration:
```bash
psql -d gator < sql/schema/002_posts.sql
```

**2. Define your SQL queries** in `sql/queries/`:

Create or update query files (e.g., `sql/queries/posts.sql`):

```sql
-- name: CreatePost :one
INSERT INTO posts (id, user_id, title, content, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPostsByUser :many
SELECT * FROM posts WHERE user_id = $1 ORDER BY created_at DESC;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;
```

**3. Generate Go code from your SQL**:

```bash
sqlc generate
```

This creates type-safe Go functions in `internal/database/` based on your SQL queries.

**4. Use the generated code** in your application:

```go
// Example usage
post, err := queries.CreatePost(ctx, db.Params{
    ID:        uuid.New(),
    UserID:    userID,
    Title:     "My Post",
    Content:   "Post content",
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
})
if err != nil {
    log.Fatal(err)
}
```

#### Understanding sqlc.yaml

The `sqlc.yaml` file configures how sqlc generates code:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "sql/queries/"
    schema: "sql/schema/"
    gen:
      go:
        out: "internal/database"
```

- **engine**: Database system (postgresql)
- **queries**: Directory containing SQL query files
- **schema**: Directory containing database schema files
- **gen**: Output configuration for generated Go code

#### Workflow for extending gator

1. Design your database changes and write SQL migrations
2. Write SQL queries in `sql/queries/`
3. Run `sqlc generate` to create Go code
4. Implement business logic using the generated functions
5. Test thoroughly before committing

This approach ensures type safety, reduces bugs, and keeps your database layer maintainable.

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
- `github.com/sqlc-dev/sqlc` - SQL code generation
