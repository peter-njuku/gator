package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
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

// Functions on posts in TUI
// --------- Post List Model ---------

type refreshPostsMsg struct{}

type postItem struct {
	title, url, feedName, publishedAt string
}

type postsModel struct {
	list         list.Model
	db           *database.Queries
	user         database.User
	state        *state
	context      context.Context
	showAddFeed  bool
	addFeedModel addFeedModel
	program      *tea.Program
}

func (i postItem) Title() string {
	return i.title
}

func (i postItem) Description() string {
	return i.feedName + " . " + i.publishedAt
}

func (i postItem) FilterValue() string {
	return i.title + " " + i.feedName
}

func (m postsModel) Init() tea.Cmd {
	return nil
}

func (m postsModel) View() string {
	if m.showAddFeed {
		return m.addFeedModel.View()
	}
	return "\n" + m.list.View() + "\nPress 'a' to add RSS feed, 'o' to open in your favourite browser, 'q' to quit\n"
}

func (m postsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showAddFeed {
		newModel, cmd := m.addFeedModel.Update(msg)
		if newModel, ok := newModel.(addFeedModel); ok {
			m.addFeedModel = newModel
		}

		if m.addFeedModel.done {
			m.showAddFeed = false
			newList, err := refreshPostsLists(m.state, &m.user)
			if err == nil {
				m.list = newList
			}
			w, h, err := getWindowSize()
			if err == nil {
				return m, func() tea.Msg { return tea.WindowSizeMsg{Width: w, Height: h} }
			}
			return m, nil
		}

		return m, cmd
	}
	switch msg := msg.(type) {
	case refreshPostsMsg:
		newLists, err := refreshPostsLists(m.state, &m.user)
		if err == nil {
			m.list = newLists
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "o":
			if len(m.list.Items()) == 0 {
				return m, nil
			}
			if selected, ok := m.list.SelectedItem().(postItem); ok {
				openURL(selected.url)
			}
			return m, nil
		case "a":
			m.showAddFeed = true
			m.addFeedModel = newAddFeedModel(m.state, &m.user)
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func newPostsModel(user database.User, s *state, program *tea.Program) (postsModel, error) {
	postsDb, err := s.db.GetPostForUser(context.Background(), user.ID)
	if err != nil {
		return postsModel{}, err
	}

	items := make([]list.Item, 0, len(postsDb))
	for _, p := range postsDb {
		items = append(items, postItem{
			title:       p.Title,
			url:         p.Url,
			feedName:    p.FeedName,
			publishedAt: p.PublishedAt.String(),
		})
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).PaddingLeft(2)
	postList := list.New(items, delegate, 80, 20)
	postList.Title = "Your Posts (press 'o' to open in browser, 'q' to quit)"
	postList.SetFilteringEnabled(true)
	m := postsModel{
		list:    postList,
		db:      s.db,
		user:    user,
		state:   s,
		context: context.Background(),
		program: program,
	}
	go m.startBackgroundScraper()
	return m, nil
}

func (m *postsModel) startBackgroundScraper() {
	if err := scrapeFeeds(m.state, m.program); err != nil {
		log.Printf("Error scraping feeds: %v\n", err)
	}
}

// Open URL in different platorms

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}

func refreshPostsLists(s *state, user *database.User) (list.Model, error) {
	postsDB, err := s.db.GetPostForUser(context.Background(), user.ID)
	if err != nil {
		return list.Model{}, err
	}

	items := make([]list.Item, 0, len(postsDB))
	for _, p := range postsDB {
		items = append(items, postItem{
			title:       p.Title,
			url:         p.Url,
			feedName:    p.FeedName,
			publishedAt: p.PublishedAt.String(),
		})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).PaddingLeft(2)
	postsLists := list.New(items, delegate, 80, 20)
	postsLists.Title = "Your Posts (press 'a' to add feed, 'o' to open in  your favourite browser, 'q' to quit)"
	postsLists.SetFilteringEnabled(true)

	return postsLists, nil
}

// Get window Size
func getWindowSize() (width, height int, err error) {
	return term.GetSize(os.Stdout.Fd())
}
