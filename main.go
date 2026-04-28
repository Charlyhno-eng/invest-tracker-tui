package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PortfolioAsset struct {
	Name   string  `json:"name"`
	Symbol string  `json:"symbol"`
	Type   string  `json:"type"`
	Shares float64 `json:"shares"`
}

type Account struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Config struct {
	Portfolio struct {
		Stocks  []PortfolioAsset `json:"stocks"`
		Cryptos []PortfolioAsset `json:"cryptos"`
	} `json:"portfolio"`
	Accounts []Account `json:"accounts"`
}

type Quote struct {
	Name          string
	Symbol        string
	Type          string
	Shares        float64
	Price         float64
	ChangePercent float64
	Value         float64
}

type quotesMsg struct {
	Quotes         []Quote
	PortfolioTotal float64
	AccountsTotal  float64
	Total          float64
	Err            error
}

type model struct {
	config         Config
	quotes         []Quote
	portfolioTotal float64
	accountsTotal  float64
	total          float64
	err            error
	loading        bool
	width          int
	height         int
}

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency           string  `json:"currency"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				PreviousClose      float64 `json:"previousClose"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
			} `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}


var (
	bgStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Background(lipgloss.Color("#F5F3FF")).
		Padding(0, 1).
		MarginBottom(1)

	cardStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#D4D4D8")).Padding(1, 2)
	kpiTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	kpiValueStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#71717A"))
	posStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#15803D")).Bold(true)
	negStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#BE123C")).Bold(true)
)


func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	m := model{
		config:  cfg,
		loading: true,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		os.Exit(1)
	}
}


func (m model) Init() tea.Cmd {
	return fetchQuotesCmd(m.config)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, fetchQuotesCmd(m.config)
		}
		return m, nil

	case quotesMsg:
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

func (m model) View() string {
	if m.loading {
		return bgStyle.Render("Loading…")
	}

	if m.err != nil {
		return bgStyle.Render(fmt.Sprintf("Error: %v\n\nr : refresh • q : quit", m.err))
	}

	header := titleStyle.Render("Invest Tracker TUI")

	kpi1 := cardStyle.Width(24).Render(
		kpiTitleStyle.Render("Portfolio") + "\n" +
			kpiValueStyle.Render(formatEUR(m.portfolioTotal)),
	)

	kpi2 := cardStyle.Width(24).Render(
		kpiTitleStyle.Render("Accounts") + "\n" +
			kpiValueStyle.Render(formatEUR(m.accountsTotal)),
	)

	kpi3 := cardStyle.Width(24).Render(
		kpiTitleStyle.Render("Total") + "\n" +
			kpiValueStyle.Render(formatEUR(m.total)),
	)

	kpis := lipgloss.JoinHorizontal(lipgloss.Top, kpi1, " ", kpi2, " ", kpi3)

	table := renderQuotesTable(m.quotes, m.width)
	accounts := renderAccounts(m.config.Accounts)
	footer := mutedStyle.Render("r : refresh • q : quit")

	content := lipgloss.JoinVertical(lipgloss.Left, header, kpis, "",table, "", accounts, "", footer)

	return bgStyle.Render(content)
}

func loadConfig(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func fetchQuotesCmd(cfg Config) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		eurUSD, err := fetchEURUSD(ctx)
		if err != nil {
			return quotesMsg{Err: err}
		}

		assets := append([]PortfolioAsset{}, cfg.Portfolio.Stocks...)
		assets = append(assets, cfg.Portfolio.Cryptos...)

		quotes := make([]Quote, 0, len(assets))
		for _, asset := range assets {
			q, err := fetchYahooQuote(ctx, asset, eurUSD)
			if err != nil {
				return quotesMsg{Err: fmt.Errorf("%s: %w", asset.Symbol, err)}
			}
			quotes = append(quotes, q)
		}

		sort.Slice(quotes, func(i, j int) bool {
			return quotes[i].Name < quotes[j].Name
		})

		var portfolioTotal float64
		for _, q := range quotes {
			portfolioTotal += q.Value
		}

		var accountsTotal float64
		for _, a := range cfg.Accounts {
			accountsTotal += a.Value
		}

		return quotesMsg{
			Quotes:         quotes,
			PortfolioTotal: portfolioTotal,
			AccountsTotal:  accountsTotal,
			Total:          portfolioTotal + accountsTotal,
		}
	}
}

func fetchEURUSD(ctx context.Context) (float64, error) {
	meta, err := fetchYahooMeta(ctx, "EURUSD=X")
	if err != nil {
		return 0, err
	}
	if meta.RegularMarketPrice == 0 {
		return 0, fmt.Errorf("EURUSD=X returned zero")
	}
	return meta.RegularMarketPrice, nil
}

func fetchYahooQuote(ctx context.Context, asset PortfolioAsset, eurUSD float64) (Quote, error) {
	meta, err := fetchYahooMeta(ctx, asset.Symbol)
	if err != nil {
		return Quote{}, err
	}

	price := meta.RegularMarketPrice
	prev := meta.PreviousClose
	if prev == 0 {
		prev = meta.ChartPreviousClose
	}
	if prev == 0 {
		return Quote{}, fmt.Errorf("previous close unavailable")
	}

	cryptoSymbols := []string{"BTC-USD", "ETH-USD", "SOL-USD"}
	if contains(cryptoSymbols, asset.Symbol) {
		price = price / eurUSD
		prev = prev / eurUSD
	}

	changePct := ((price - prev) / prev) * 100

	return Quote{
		Name:          asset.Name,
		Symbol:        asset.Symbol,
		Type:          asset.Type,
		Shares:        asset.Shares,
		Price:         round(price, 2),
		ChangePercent: round(changePct, 2),
		Value:         round(price*asset.Shares, 2),
	}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func fetchYahooMeta(ctx context.Context, symbol string) (*struct {
	Currency           string  `json:"currency"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	PreviousClose      float64 `json:"previousClose"`
	ChartPreviousClose float64 `json:"chartPreviousClose"`
}, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1d", symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload yahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Chart.Result) == 0 {
		return nil, fmt.Errorf("no result for %s", symbol)
	}

	return &payload.Chart.Result[0].Meta, nil
}

func renderQuotesTable(quotes []Quote, width int) string {
	var lines []string

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerStyle.Width(40).Render("Asset"),
		headerStyle.Width(12).Render("Price"),
		headerStyle.Width(10).Render("Var %"),
		headerStyle.Width(12).Render("Parts"),
		headerStyle.Width(14).Render("Value"),
	)
	lines = append(lines, header)

	for _, q := range quotes {
		change := formatPercent(q.ChangePercent)
		if q.ChangePercent >= 0 {
			change = posStyle.Width(10).Render(change)
		} else {
			change = negStyle.Width(10).Render(change)
		}

		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(40).Render(truncate(q.Name, 40)),
			lipgloss.NewStyle().Width(12).Render(formatEUR(q.Price)),
			change,
			lipgloss.NewStyle().Width(12).Render(formatShares(q.Shares)),
			lipgloss.NewStyle().Width(14).Render(formatEUR(q.Value)),
		)
		lines = append(lines, line)
	}

	return cardStyle.Render(strings.Join(lines, "\n"))
}

func renderAccounts(accounts []Account) string {
	var lines []string
	lines = append(lines, headerStyle.Render("Accounts"))

	for _, a := range accounts {
		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(40).Render(a.Name),
			lipgloss.NewStyle().Width(14).Render(formatEUR(a.Value)),
		)
		lines = append(lines, line)
	}

	return cardStyle.Render(strings.Join(lines, "\n"))
}

func formatEUR(v float64) string {
	return fmt.Sprintf("%.2f €", v)
}

func formatPercent(v float64) string {
	if v > 0 {
		return fmt.Sprintf("+%.2f%%", v)
	}
	return fmt.Sprintf("%.2f%%", v)
}

func formatShares(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.4f", v)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

func round(v float64, places int) float64 {
	pow := 1.0
	for i := 0; i < places; i++ {
		pow *= 10
	}
	if v >= 0 {
		return float64(int(v*pow+0.5)) / pow
	}
	return float64(int(v*pow-0.5)) / pow
}
