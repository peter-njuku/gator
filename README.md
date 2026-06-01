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

## Command Reference

### Authentication

**`register <username>`**

Create a new user account. Prompts for password interactively (input is hidden). Password confirmation is required.

```bash
gator register alice
# (prompts for password and confirmation)
```

**`login <username>`**

Log in as an existing user. Prompts for password interactively (input is hidden). Sets the username in your config for subsequent commands.

Note: Press Ctrl+C to cancel safely without breaking terminal echo.

```bash
gator login alice
# (prompts for password)
```

### User Management

**`users`**

List all registered users in the database.

```bash
gator users
```

**`reset`**

Reset the database by deleting all users and associated data. Use with caution.

```bash
gator reset
```

### Feed Management

**`addfeed <feed_name> <feed_url>`**

Add a new feed and automatically follow it. Requires being logged in. Both name and URL are required.

```bash
gator addfeed "Hacker News" "https://news.ycombinator.com/rss"
```

**`feeds`**

List all feeds in the database (regardless of whether you follow them).

```bash
gator feeds
```

**`follow <feed_url>`**

Follow an existing feed by URL. Requires being logged in.

```bash
gator follow "https://news.ycombinator.com/rss"
```

**`following`**

List all feeds the currently logged-in user is following.

```bash
gator following
```

**`unfollow <feed_url>`**

Stop following a feed. Requires being logged in.

```bash
gator unfollow "https://news.ycombinator.com/rss"
```

### Feed Aggregation

**`agg <duration>`**

Run the feed aggregator in a continuous loop. Fetches new posts from all feeds at the specified interval.

Duration format: `10s`, `1m`, `5m30s`, etc.

```bash
gator agg 30s
# Fetches feeds every 30 seconds (runs until Ctrl+C)

gator agg 1m
# Fetches feeds every 1 minute
```

### Browsing Posts

**`browse [limit] [--feed <feed_filter>] [--sort {asc|desc}]`**

Browse recent posts from followed feeds. Requires being logged in.

Options:
- `limit` (optional, default: 2) — Number of posts to display
- `--feed <filter>` — Filter posts by feed name (case-insensitive substring matching)
- `--sort {asc|desc}` (default: desc) — Sort by publication date; `asc` shows oldest first, `desc` shows newest first

```bash
# Show 2 most recent posts
gator browse

# Show 10 most recent posts
gator browse 10

# Show 5 posts sorted by oldest first
gator browse 5 --sort asc

# Show 10 posts from feeds matching "news" (case-insensitive)
gator browse 10 --feed news

# Combine filters: 20 posts from "golang" feeds, oldest first
gator browse 20 --feed golang --sort asc
```

### Notes on Commands

- Commands requiring login (addfeed, follow, unfollow, following, browse) will fail if no user is set in `~/.gatorconfig.json`
- Password entry is fully interactive and handles Ctrl+C safely without breaking your terminal
- Feed name filtering in `browse` uses case-insensitive substring matching (e.g., `--feed python` matches "Python Weekly" and "python-news")

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

# Problems I found
1. On login, if the user cancels the program on the enter password phase, the echo capabilities of my 
terminal would not work. I had to ask my Chinese friend deepseek. This helped and I had to handle 
some syscall command. I did not anticipate for this but it finally worked.

```go
ar passwordBytes []byte
	for {
		var b [1]byte
		_, err := os.Stdin.Read(b[:])
		if err != nil {
			return "", err
		}

		//ctrl + c
		if b[0] == 0x03 {
			return "", fmt.Errorf("Interrupted")
		}

		// Enter or Return
		if b[0] == '\r' || b[0] == '\n' {
			break
		}

		//Deleting Backspace
		if b[0] == 127 || b[0] == 8 {
			if len(passwordBytes) > 0 {
				passwordBytes = passwordBytes[:len(passwordBytes)-1]
				fmt.Print("\b \b")
			}
			continue
		}

		passwordBytes = append(passwordBytes, b[0])
		fmt.Print("")
	}

	fmt.Println()
  ```
Thank you [Deepseek](https://chat.deepseek.com/)
```