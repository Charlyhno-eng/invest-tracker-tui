package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"invest-tracker-tui/internal/config"
	"invest-tracker-tui/internal/fetch"
	"invest-tracker-tui/internal/utils"
)


type Model struct {
	config         config.Config
	quotes         []fetch.Quote
	portfolioTotal float64
	accountsTotal  float64
	total          float64
	err            error
	loading        bool
	width          int
	height         int
}


func NewModel(cfg config.Config) Model {
	return Model{
		config:  cfg,
		loading: true,
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
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			m.err = nil
			return m, fetch.QuotesCmd(m.config)
		}
		return m, nil

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
	}

	return m, nil
}

func (m Model) View() string {
	if m.loading {
		return bgStyle.Render("Chargement…")
	}

	if m.err != nil {
		return bgStyle.Render("Erreur : " + m.err.Error() + "\n\nr : rafraîchir • q : quitter")
	}

	header := titleStyle.Render("Invest Tracker TUI")

	kpi1 := renderKPI("Portefeuille", utils.FormatEUR(m.portfolioTotal), 24)
	kpi2 := renderKPI("Comptes", utils.FormatEUR(m.accountsTotal), 24)
	kpi3 := renderKPI("Total", utils.FormatEUR(m.total), 24)
	kpis := lipgloss.JoinHorizontal(lipgloss.Top, kpi1, " ", kpi2, " ", kpi3)

	table := renderQuotesTable(m.quotes)
	accounts := renderAccounts(m.config.Accounts)
	footer := mutedStyle.Render("r : rafraîchir • q : quitter")

	content := lipgloss.JoinVertical(lipgloss.Left,
		header, kpis, "", table, "", accounts, "", footer,
	)

	return bgStyle.Render(content)
}
