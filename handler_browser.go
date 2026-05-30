package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/peter-njuku/gator/internal/database"
)

func handlerBrowser(s *state, cmd command, user database.User) error {
	limit := int32(2)

	if len(cmd.Args) > 0 {
		customLimit, err := strconv.ParseInt(cmd.Args[0], 32, 10)
		if err != nil {
			return fmt.Errorf("Invalid limit: %w", err)
		}
		limit = int32(customLimit)
	}

	posts, err := s.db.GetPostForUser(context.Background(), database.GetPostForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("Could not get post for user: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts in here, try making some by scraping your favourite sites")
		return nil
	}

	fmt.Printf("\n📰 Recent posts for %s (showing %d of %d):\n", user.Name, len(posts), limit)
	fmt.Println("=====================================")

	for i, post := range posts {
		fmt.Printf("\n%d. %s\n", i+1, post.Title)
		fmt.Printf("   📌 Feed: %s\n", post.FeedName)
		fmt.Printf("   📅 Published: %s\n", post.PublishedAt)
		fmt.Printf("   🔗 URL: %s\n", post.Url)

		if post.Description.String != "" {
			desc := post.Description.String
			if len(desc) > 200 {
				desc = desc[:200] + "..."
			}

			fmt.Printf("   📝 Description: %s\n", desc)
		}

		fmt.Println("   ---")
	}
	fmt.Printf("\nTotal posts shown: %d\n", len(posts))
	if int32(len(posts)) == limit {
		fmt.Printf("\n💡 Tip: To see more posts, run: browse %d\n", limit+5)
	}

	return nil
}
