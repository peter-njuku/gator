package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/google/uuid"
	"github.com/peter-njuku/gator/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func HandlerLogin(s *state, cmd command) error {

	if len(cmd.Args) == 0 {
		return fmt.Errorf("Usage: %v <name>", cmd.Name)
	}

	username := cmd.Args[0]
	password, err := readPassword("Password: ")
	if err != nil {
		return err
	}

	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("User not found")
		}
		return fmt.Errorf("Could not find user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return fmt.Errorf("Invalid password")
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("Could not set current user: %w", err)
	}

	fmt.Println("User switched successfully")
	fmt.Println("Logged in as:", username)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("Usage: %v <name>", cmd.Name)
	}

	name := cmd.Args[0]

	password, err := readPassword("Enter your password: ")
	if err != nil {
		return fmt.Errorf("Could not read your password")
	}
	confirm, err := readPassword("Confirm your password: ")
	if err != nil {
		return fmt.Errorf("Could not read your confirmation")
	}

	if password != confirm {
		return fmt.Errorf("Passwords do not match")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Could not hash password: %w", err)
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:           uuid.New(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Name:         name,
		PasswordHash: string(hashed),
	})
	if err != nil {
		return fmt.Errorf("Could not create user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Could not set user: %w", err)
	}

	fmt.Println("User created and set successfully")
	printUser(user)

	return nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	fd := os.Stdin.Fd()
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}

	defer term.Restore(fd, oldState)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	defer signal.Stop(sigChan)

	var passwordBytes []byte
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

	return string(passwordBytes), nil
}

func printUser(user database.User) {
	fmt.Printf(" * ID:      %v\n", user.ID)
	fmt.Printf(" * Name:    %v\n", user.Name)
}
