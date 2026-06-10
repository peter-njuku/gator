package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/peter-njuku/gator/internal/database"
)

func scrapeFeeds(s *state, program *tea.Program) error {
	for {
		feed, err := s.db.GetNextFeedToFetch(context.Background())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				time.Sleep(10 * time.Second)
				continue
			}
			return fmt.Errorf("Error getting next feed: %w", err)
		}

		log.Printf("Fetching feed: %s\n", feed.Name)
		rssFeed, err := fetchFeed(context.Background(), feed.Url)
		if err != nil {
			log.Printf("Error fetching feed: %s (%s)", feed.Name, feed.Url)
			err = s.db.MarkFeedFetched(context.Background(), feed.ID)
			if err != nil {
				return fmt.Errorf("Could not mark feed as fetched: %w", err)
			}
			time.Sleep(10 * time.Second)
			continue
		}

		err = s.db.MarkFeedFetched(context.Background(), feed.ID)
		if err != nil {
			return fmt.Errorf("Could not mark feed as fetched: %w", err)
		}

		log.Printf("Processing %d posts from %s....\n", len(rssFeed.Channel.Item), feed.Name)

		for _, item := range rssFeed.Channel.Item {
			if item.Title == "" || item.Link == "" {
				log.Println("Skipping post with no title ot link...")
				continue
			}

			publishedAt, err := parsePublishedAt(item.PubDate)
			if err != nil {
				log.Printf("Warning: %s for post %s", err, item.Title)
				publishedAt = time.Now().UTC()
			}

			err = s.db.CreatePost(context.Background(), database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Title:       item.Title,
				Url:         item.Link,
				Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
				PublishedAt: publishedAt,
				FeedID:      feed.ID,
			})

			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
					continue
				}
				log.Printf("Error saving post '%s' : %v\n", item.Title, err)
			}
		}
		log.Printf("Successfully processed feed: %s\n", feed.Name)
		log.Println("=====================================")

		if program != nil {
			program.Send(refreshPostsMsg{})
		}

		time.Sleep(10 * time.Second)
	}
}

func parsePublishedAt(pubDate string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05 -0700 MST",
		"Mon, 02 Jan 2006 15:04:05 GMT",
		"Mon, 02 Jan 2006 15:04:05 +0000",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, pubDate); err == nil {
			return t.UTC(), nil
		}
	}

	if timestamp, err := strconv.ParseInt(pubDate, 10, 64); err == nil {
		return time.Unix(timestamp, 0).UTC(), nil
	}

	return time.Time{}, fmt.Errorf("Unable to parse date: %s", pubDate)
}
