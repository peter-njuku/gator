package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/peter-njuku/gator/internal/database"
)

func handlerBrowser(s *state, cmd command, user database.User) error {
	browseFlags := flag.NewFlagSet("browse", flag.ContinueOnError)
	var (
		feedName  string
		sortOrder string
	)

	browseFlags.StringVar(&feedName, "feed", "", "Filter by feed name (exact match, case-insensitive)")
	browseFlags.StringVar(&sortOrder, "sort", "desc", "Sort order: asc (oldest first), or desc(newest first)")

	err := browseFlags.Parse(cmd.Args)
	if err != nil {
		return fmt.Errorf("Error parsing flags: %w", err)
	}

	remaining := browseFlags.Args()

	limit := int32(2)
	if len(remaining) > 0 {
		n, err := strconv.ParseInt(remaining[0], 10, 32)
		if err != nil {
			return fmt.Errorf("Invalid limit: %w", err)
		}
		limit = int32(n)
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		return fmt.Errorf("Sort order must be 'asc' or 'desc'")
	}

	dbPosts, err := s.db.GetPostForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Could not get post for user: %w", err)
	}

	type displayPost struct {
		Title       string
		FeedName    string
		PublishedAt time.Time
		Url         string
		Description string
	}

	var posts []displayPost

	for _, post := range dbPosts {

		posts = append(posts, displayPost{
			Title:       post.Title,
			FeedName:    post.FeedName,
			PublishedAt: post.PublishedAt,
			Url:         post.Url,
			Description: post.Description.String,
		})
	}

	if feedName != "" {
		lowerFeedName := strings.ToLower(feedName)
		filtered := []displayPost{}
		for _, post := range posts {
			if strings.Contains(strings.ToLower(post.FeedName), lowerFeedName) {
				filtered = append(filtered, post)
			}
		}
		posts = filtered
	}

	sort.Slice(posts, func(i, j int) bool {
		if sortOrder == "asc" {
			return posts[i].PublishedAt.Before(posts[j].PublishedAt)
		}
		return posts[i].PublishedAt.After(posts[j].PublishedAt)
	})

	if int(limit) < len(posts) {
		posts = posts[:limit]
	}

	if len(posts) == 0 {
		fmt.Println("No posts in here, try making some by scraping your favourite sites")
		return nil
	}

	fmt.Printf(" - showing %d posts (sort: %s):\n", len(posts), sortOrder)
	fmt.Println("=====================================")

	for i, post := range posts {
		fmt.Printf("\n%d. %s\n", i+1, post.Title)
		fmt.Printf("   📌 Feed: %s\n", post.FeedName)
		fmt.Printf("   📅 Published: %s\n", post.PublishedAt)
		fmt.Printf("   🔗 URL: %s\n", post.Url)

		if post.Description != "" {
			desc := post.Description
			if len(desc) > 200 {
				desc = desc[:200] + "..."
			}

			fmt.Printf("   📝 Description: %s\n", desc)
		}

		fmt.Println("   ---")
	}
	fmt.Printf("\nTotal posts shown: %d\n", len(posts))
	if int32(len(posts)) == limit && limit > 0 {
		fmt.Printf("\n💡 Tip: To see more posts, run: browse %d\n", limit+5)
	}

	if feedName == "" {
		fmt.Println("💡 Tip: Filter by feed: browse --feed=<feed_name>")
	}

	fmt.Println("💡 Tip: Change sort order: browse --sort=asc")

	return nil
}
