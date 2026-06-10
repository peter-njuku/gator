package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

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
		postsUI, err := newPostsModel(*regResult.user, s, nil)
		if err != nil {
			return fmt.Errorf("Could not diplay posts: %w", err)
		}
		p = tea.NewProgram(postsUI, tea.WithAltScreen())
		postsUI.program = p
		_, err = p.Run()
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

	postsUI, err := newPostsModel(*loginResult.user, s, nil)
	if err != nil {
		return fmt.Errorf("Could not diplay posts: %w", err)
	}
	p = tea.NewProgram(postsUI, tea.WithAltScreen())
	postsUI.program = p
	_, err = p.Run()

	return err
}
