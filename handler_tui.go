package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/peter-njuku/gator/internal/database"
	"golang.org/x/crypto/bcrypt"
)

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

func handlerTui(s *state, cmd command) error {
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
	return nil
}
