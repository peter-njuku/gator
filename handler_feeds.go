package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/peter-njuku/gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed
	err = xml.Unmarshal(dat, &rssFeed)
	if err != nil {
		return nil, err
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i, item := range rssFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		rssFeed.Channel.Item[i] = item
	}

	return &rssFeed, nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		fmt.Printf("Usage: %s <name> <url>", cmd.Name)
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

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Feed created but could not follow: %w", err)
	}

	fmt.Println("Feed create successfully")
	printFeed(feed, user)
	fmt.Println("\nYou are now following this feed")
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

// Functions of TUI - Feeds
type addFeedModel struct {
	inputs     []textinput.Model
	focusIndex int
	errMsg     string
	state      *state
	user       *database.User
	done       bool
}

func (m addFeedModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addFeedModel) View() string {
	var b strings.Builder
	b.WriteString("Add a new RSS feed\n\n")
	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	b.WriteString("\n\n")

	if m.errMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗" + m.errMsg + "\n"))
	}
	b.WriteString("(Tab to switch, Enter/Return to submit, Ctrl+c to quit)")

	return b.String()
}

func (m addFeedModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	cmd := tea.Batch(cmds...)
	return cmd
}

func newAddFeedModel(s *state, user *database.User) addFeedModel {
	m := addFeedModel{
		inputs: make([]textinput.Model, 2),
		user:   user,
		state:  s,
	}

	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Feed name"
	m.inputs[0].Focus()

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Feed URL"

	return m
}

func (m addFeedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				name := m.inputs[0].Value()
				url := m.inputs[1].Value()

				if name == "" || url == "" {
					m.errMsg = "Both fields are required"
					return m, nil
				}

				feed, err := m.state.db.CreateFeed(context.Background(), database.CreateFeedParams{
					ID:        uuid.New(),
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					Name:      name,
					Url:       url,
					UserID:    m.user.ID,
				})
				if err != nil {
					_, err := m.state.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
						UserID: m.user.ID,
						FeedID: feed.ID,
					})
					if err != nil {
						m.errMsg = "Could not create follow for the feed" + err.Error()
					} else {
						m.errMsg = "Could not create RSS Feed. TRy again later"
						return m, nil
					}
				}

				_, err = m.state.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
					UserID: m.user.ID,
					FeedID: feed.ID,
				})
				if err != nil {
					m.errMsg = "Could not Auto-Follow the feed" + err.Error()
					return m, nil
				}
				err = handlerAggTUI(m.state, nil)
				if err != nil {
					m.errMsg = "HAnlder AgG is FuCked like sHiT"
					return m, nil
				}
				m.done = true
				return m, nil
			}

			m.inputs[m.focusIndex].Blur()
			m.focusIndex++
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case "tab", "shift+tab":
			m.inputs[m.focusIndex].Blur()
			if msg.String() == "tab" {
				m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			} else {
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs) - 1
				}
			}

			m.inputs[m.focusIndex].Focus()
			return m, nil
		}
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}
