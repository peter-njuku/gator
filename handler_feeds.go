package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/peter-njuku/gator/internal/database"
)

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.Args) != 2 {
		fmt.Printf("Usage: %s <name> <url>", cmd.Name)
	}

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUsername)
	if err != nil {
		return fmt.Errorf("Could not get user: %w", err)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})

	if err != nil {
		return fmt.Errorf("Could not create feed: %w", err)
	}

	fmt.Println("Feed create successfully")
	printFeed(feed, user)
	fmt.Println("=====================================\nEOF\n=====================================")
	return nil
}

func handlerAllFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Could ot fetch feeds: %w", err)
	}

	for _, feed := range feeds {
		user, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("COuld not get user name: %w", err)
		}
		printFeed(feed, user)
		fmt.Println("=====================================\nEOF\n=====================================")
	}
	return nil
}

func printFeed(feed database.Feed, user database.User) {
	fmt.Printf("* ID:            	%s\n", feed.ID)
	fmt.Printf("* Created:       	%v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       	%v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          	%s\n", feed.Name)
	fmt.Printf("* URL:           	%s\n", feed.Url)
	fmt.Printf("* UserID:        	%s\n", feed.UserID)
	fmt.Printf("* Owners Name:      	%s\n", user.Name)
}
