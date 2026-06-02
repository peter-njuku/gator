package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/peter-njuku/gator/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// --------- Register Model ---------

type registerModel struct {
	inputs        []textinput.Model
	focusIndex    int
	errMsg        string
	state         *state
	readyToSwitch bool
	user          *database.User
}

func (m registerModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m registerModel) View() string {
	var b strings.Builder
	b.WriteString("Register to Gator\n\n")
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

func (m registerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				username := m.inputs[0].Value()
				password := m.inputs[1].Value()
				confirm := m.inputs[2].Value()

				if username == "" {
					m.errMsg = "Username cannot be empty"
					return m, nil
				}

				if password == "" {
					m.errMsg = "Password cannot be empty"
					return m, nil
				}

				if password != confirm {
					m.errMsg = "Password Mismatch"
					return m, nil
				}

				_, err := m.state.db.GetUser(context.Background(), username)
				if err == nil {
					m.errMsg = "Username already exists"
					return m, nil
				} else if !errors.Is(err, sql.ErrNoRows) {
					m.errMsg = "Database error"
					return m, nil
				}

				// Hash password
				hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					m.errMsg = "Could not hash password"
					return m, nil
				}

				user, err := m.state.db.CreateUser(context.Background(), database.CreateUserParams{
					ID:           uuid.New(),
					CreatedAt:    time.Now().UTC(),
					UpdatedAt:    time.Now().UTC(),
					Name:         username,
					PasswordHash: string(hashed),
				})
				if err != nil {
					m.errMsg = "Could not create user"
					return m, nil
				}

				//success
				m.user = &user
				m.readyToSwitch = true
				return m, tea.Quit
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

func (m *registerModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	cmd := tea.Batch(cmds...)
	return cmd
}

func newRegisterModel(s *state) registerModel {
	m := registerModel{
		inputs: make([]textinput.Model, 3),
		state:  s,
	}

	// Username input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Enter new username"
	m.inputs[0].Focus()

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Enter new password"
	m.inputs[1].EchoMode = textinput.EchoPassword
	m.inputs[1].EchoCharacter = '*'

	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Confirm your password"
	m.inputs[2].EchoMode = textinput.EchoPassword
	m.inputs[2].EchoCharacter = '*'

	return m
}

// --------- Login Model ---------
type loginModel struct {
	inputs        []textinput.Model
	focusIndex    int
	errMsg        string
	state         *state
	readyToSwitch bool
	user          *database.User
}

func (m loginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m loginModel) View() string {
	var b strings.Builder
	b.WriteString("Login to Gator\n\n")
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

func (m loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				username := m.inputs[0].Value()
				password := m.inputs[1].Value()

				user, err := m.state.db.GetUser(context.Background(), username)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						m.errMsg = "User not found"
					} else {
						m.errMsg = "Database error"
					}
					return m, nil
				}

				err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
				if err != nil {
					m.errMsg = "Invalid password"
					return m, nil
				}

				//success
				m.user = &user
				m.readyToSwitch = true
				return m, tea.Quit
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

func (m *loginModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	cmd := tea.Batch(cmds...)
	return cmd
}

func newLoginModel(s *state) loginModel {
	m := loginModel{
		inputs: make([]textinput.Model, 2),
		state:  s,
	}

	// Username input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "username"
	m.inputs[0].Focus()

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "password"
	m.inputs[1].EchoMode = textinput.EchoPassword
	m.inputs[1].EchoCharacter = '*'

	return m
}

// --------- Post List Model ---------

type postItem struct {
	title, url, feedName, publishedAt string
}

type postsModel struct {
	list    list.Model
	db      *database.Queries
	user    database.User
	context context.Context
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
	return "\n" + m.list.View() + "\nPress 'o' to open in your favourite browser, 'q' to quit\n"
}

func (m postsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

func newPostsModel(user database.User, s *state) (postsModel, error) {
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
	return postsModel{
		list:    postList,
		db:      s.db,
		user:    user,
		context: context.Background(),
	}, nil
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

// --------- Handler Function for Terminal User Interface ---------
func handlerTui(s *state, cmd command) error {
	if len(cmd.Args) > 0 && cmd.Args[0] == "register" {
		reg := newRegisterModel(s)

		p := tea.NewProgram(reg, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		regResult := finalModel.(registerModel)
		if !regResult.readyToSwitch {
			return nil
		}
		postsUI, err := newPostsModel(*regResult.user, s)
		if err != nil {
			return fmt.Errorf("Could not diplay posts: %w", err)
		}
		_, err = tea.NewProgram(postsUI, tea.WithAltScreen()).Run()
		return err
	}
	login := newLoginModel(s)
	p := tea.NewProgram(login, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	loginResult := finalModel.(loginModel)
	if !loginResult.readyToSwitch {
		return nil
	}

	postsUI, err := newPostsModel(*loginResult.user, s)
	if err != nil {
		return fmt.Errorf("Could not diplay posts: %w", err)
	}
	_, err = tea.NewProgram(postsUI, tea.WithAltScreen()).Run()
	return err
}
