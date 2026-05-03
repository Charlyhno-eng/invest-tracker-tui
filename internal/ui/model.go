package ui

import (
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"invest-tracker-tui/internal/config"
	"invest-tracker-tui/internal/fetch"
)

const yahooURL = "https://fr.finance.yahoo.com/"


type appView int

const (
	viewDashboard appView = iota
	viewWatchlist
)

type Model struct {
	config    config.Config
	watchlist config.Watchlist

	// Dashboard state.
	quotes         []fetch.Quote
	portfolioTotal float64
	accountsTotal  float64
	total          float64
	loading        bool
	err            error

	// Watchlist state.
	watchSections  []fetch.WatchSection
	watchLoading   bool
	watchErr       error

	view   appView
	width  int
	height int
}


// NewModel returns an initial Model ready to fetch quotes.
func NewModel(cfg config.Config, wl config.Watchlist) Model {
	return Model{
		config:    cfg,
		watchlist: wl,
		loading:   true,
	}
}

func (m Model) Init() tea.Cmd {
	return fetch.QuotesCmd(m.config)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg.String())

	case fetch.QuotesMsg:
		m.loading = false
		m.err = msg.Err
		if msg.Err == nil {
			m.quotes = msg.Quotes
			m.portfolioTotal = msg.PortfolioTotal
			m.accountsTotal = msg.AccountsTotal
			m.total = msg.Total
		}
		return m, nil

	case fetch.WatchlistMsg:
		m.watchLoading = false
		m.watchErr = msg.Err
		if msg.Err == nil {
			m.watchSections = msg.Sections
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch m.view {
	case viewWatchlist:
		switch key {
		case "esc", "b":
			m.view = viewDashboard
			m.watchErr = nil
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.watchLoading = true
			m.watchErr = nil
			return m, fetch.WatchlistCmd(m.watchlist)
		}
		return m, nil

	default: // viewDashboard
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			m.err = nil
			return m, fetch.QuotesCmd(m.config)
		case "y":
			return m, openBrowserCmd(yahooURL)
		case "w":
			m.view = viewWatchlist
			if m.watchSections == nil {
				m.watchLoading = true
				return m, fetch.WatchlistCmd(m.watchlist)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.view {
	case viewWatchlist:
		return m.renderWatchlistView()
	default:
		return m.renderDashboardView()
	}
}

func (m Model) renderDashboardView() string {
	if m.loading {
		return bgStyle.Render(renderLoading())
	}
	if m.err != nil {
		return bgStyle.Render(renderError(m.err))
	}

	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorBorder)).
		Render("")

	content := lipgloss.JoinVertical(lipgloss.Left,
		renderHeader(m.width),
		"",
		renderKPIs(m.portfolioTotal, m.accountsTotal, m.total),
		"",
		sep,
		renderPortfolioTable(m.quotes),
		"",
		renderAccounts(m.config.Accounts),
		"",
		renderFooter(),
	)

	return bgStyle.Render(content)
}

func (m Model) renderWatchlistView() string {
	if m.watchLoading {
		return bgStyle.Render(renderLoading())
	}
	if m.watchErr != nil {
		return bgStyle.Render(renderError(m.watchErr))
	}
	return bgStyle.Render(renderWatchlist(m.watchSections, m.width))
}

// openBrowserCmd opens the given URL in the default system browser.
func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		_ = cmd.Start()
		return nil
	}
}
