package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/peter-njuku/gator/internal/config"
	"github.com/peter-njuku/gator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Unable to Read config file: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatalf("Unable to cennect to database: %v", err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	s := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	if len(os.Args) < 2 {
		log.Fatal("No command provided")
	}

	cmd := command{
		Name: os.Args[1],
		Args: []string(os.Args[2:]),
	}

	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	cmds.register("login", HandlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetAllUsers)

	if err := cmds.run(s, cmd); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}
