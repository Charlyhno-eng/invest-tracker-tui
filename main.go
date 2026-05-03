package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"invest-tracker-tui/internal/config"
	"invest-tracker-tui/internal/ui"
)

func main() {
	cfg, err := config.Load("assets/portfolio.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "portfolio error: %v\n", err)
		os.Exit(1)
	}

	wl, err := config.LoadWatchlist("assets/watchlist.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "watchlist error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel(cfg, wl), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		os.Exit(1)
	}
}
