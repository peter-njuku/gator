package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/peter-njuku/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currentUserName := s.cfg.CurrentUsername
		if currentUserName == "" {
			return fmt.Errorf("You must be logged in to run this command")
		}

		user, err := s.db.GetUser(context.Background(), currentUserName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("User %s not found in database", currentUserName)
			}
			return fmt.Errorf("Could not fetch user: %w", err)
		}
		return handler(s, cmd, user)
	}
}
