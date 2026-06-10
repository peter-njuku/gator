package main

import (
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return errors.New("Argument for time between requests should be provided")
	}
	time_btn_req := cmd.Args[0]
	duration, err := time.ParseDuration(time_btn_req)
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %s", time_btn_req)
	tick := time.NewTicker(duration)
	for range tick.C {
		if err := scrapeFeeds(s, nil); err != nil {
			return err
		}
	}
	return nil
}

func handlerAggTUI(s *state, program *tea.Program) error {
	go func() {
		if err := scrapeFeeds(s, program); err != nil {
			fmt.Printf("Error scraping feeds: %v", err)
		}
	}()
	return nil
}
