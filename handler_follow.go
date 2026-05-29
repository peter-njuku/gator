package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/peter-njuku/gator/internal/database"
)

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage %s <url>", cmd.Name)
	}

	feedURL := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("Could not find feed by such URL: %s", feedURL)
		}
		return fmt.Errorf("Could not find feed: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Could not follow feed: %w", err)
	}

	fmt.Printf("%s is now following %s\n", feedFollow.UserName, feedFollow.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Could not get followed feeds: %w", err)
	}
	if len(feedFollows) == 0 {
		fmt.Printf("You are not following any feeds yet.\n")
		fmt.Printf("Try: follow <feed_url>\n")
		return nil
	}

	fmt.Printf("Feeds followed by %s:\n", user.Name)
	fmt.Println("=====================================")

	for i, follow := range feedFollows {
		fmt.Printf("%d. %s\n", i+1, follow.FeedName)
		fmt.Printf("   URL: %s\n", follow.FeedUrl)
		fmt.Printf("   Followed since: %s\n", follow.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("=====================================")
		fmt.Println()
	}

	fmt.Printf("Total feeds followed: %d\n", len(feedFollows))
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: %s <feedURL>", cmd.Name)
	}

	feedURL := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("No feed with such a URL. Try another plzzz!!!!")
		}
		return fmt.Errorf("Could not fetch feed: %w", err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Could not delete follow: %w", err)
	}
	fmt.Println("You have unfollowed", feed.Name)
	return nil
}
