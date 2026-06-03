package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func handlerGetAllUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Could not get users: %w", err)
	}

	currentUser := s.cfg.GetCurrentUser()

	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

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

		//Deleting Backspace
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

// Funtions to use in TUI
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
