package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"invest-tracker-tui/internal/config"
	"invest-tracker-tui/internal/fetch"
	"invest-tracker-tui/internal/utils"
)


// ── Header ────────────────────────────────────────────────────────────────────

func renderHeader(width int) string {
	now := time.Now().Format("01/02/2006  15:04:05")
	logo := logoStyle.Render("▸ INVEST TRACKER")
	ts := subtitleStyle.Render(now)

	gap := width - lipgloss.Width(logo) - lipgloss.Width(ts) - 6
	if gap < 1 {
		gap = 1
	}

	top := lipgloss.JoinHorizontal(lipgloss.Top,
		logo,
		strings.Repeat(" ", gap),
		ts,
	)
	line := headerLineStyle.Render(strings.Repeat("─", width-6))

	return lipgloss.JoinVertical(lipgloss.Left, top, line)
}

// ── KPI cards ─────────────────────────────────────────────────────────────────

func renderKPIs(portfolio, accounts, total float64) string {
	c1 := renderKPICard("PORTFOLIO", utils.FormatEUR(portfolio), false, 26)
	c2 := renderKPICard("ACCOUNTS", utils.FormatEUR(accounts), false, 26)
	c3 := renderKPICard("NET TOTAL", utils.FormatEUR(total), true, 26)

	return lipgloss.JoinHorizontal(lipgloss.Top, c1, "  ", c2, "  ", c3)
}

func renderKPICard(label, value string, highlight bool, width int) string {
	s := kpiCardStyle
	vs := kpiValueStyle
	if highlight {
		s = kpiCardHiStyle
		vs = kpiTotalValueStyle
	}
	return s.Width(width).Render(
		kpiLabelStyle.Render(label) + "\n" + vs.Render(value),
	)
}

// ── Portfolio table ───────────────────────────────────────────────────────────

func renderPortfolioTable(quotes []fetch.Quote) string {
	sectionTitle := sectionStyle.Render("PORTFOLIO")
	sectionLine := sectionLineStyle.Render(strings.Repeat("─", 60))
	section := lipgloss.JoinVertical(lipgloss.Left, sectionTitle, sectionLine)

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		tableHeaderStyle.Width(7).Render("TYPE"),
		tableHeaderStyle.Width(32).Render("NAME"),
		tableHeaderStyle.Width(12).Render("SYMBOL"),
		tableHeaderStyle.Width(13).Render("PRICE"),
		tableHeaderStyle.Width(10).Render("CHG %"),
		tableHeaderStyle.Width(10).Render("SHARES"),
		tableHeaderStyle.Width(14).Render("VALUE"),
	)

	sep := cellDimStyle.Render(strings.Repeat("·", 98))

	rows := []string{header, sep}
	for _, q := range quotes {
		rows = append(rows, renderQuoteRow(q))
	}

	return lipgloss.JoinVertical(lipgloss.Left, section, strings.Join(rows, "\n"))
}

func renderQuoteRow(q fetch.Quote) string {
	badge := badgeForType(q.Type)

	change := utils.FormatPercent(q.ChangePercent)
	if q.ChangePercent >= 0 {
		change = posStyle.Width(10).Render(change)
	} else {
		change = negStyle.Width(10).Render(change)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		badge,
		cellStyle.Width(32).Render(utils.Truncate(q.Name, 31)),
		symbolStyle.Width(12).Render(q.Symbol),
		cellStyle.Width(13).Render(utils.FormatEUR(q.Price)),
		change,
		cellDimStyle.Width(10).Render(utils.FormatShares(q.Shares)),
		cellStyle.Width(14).Render(utils.FormatEUR(q.Value)),
	)
}

// badgeForType returns a colored type badge for a given asset type.
func badgeForType(t config.AssetType) string {
	switch t {
	case config.AssetTypeCrypto:
		return badgeCryptoStyle.Width(7).Render("CRYPTO")
	case config.AssetTypeETF:
		return badgeEtfStyle.Width(7).Render("ETF   ")
	default:
		return badgeStockStyle.Width(7).Render("STOCK ")
	}
}

// ── Accounts ──────────────────────────────────────────────────────────────────

func renderAccounts(accounts []config.Account) string {
	sectionTitle := sectionStyle.Render("ACCOUNTS")
	sectionLine := sectionLineStyle.Render(strings.Repeat("─", 60))
	section := lipgloss.JoinVertical(lipgloss.Left, sectionTitle, sectionLine)

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		tableHeaderStyle.Width(44).Render("NAME"),
		tableHeaderStyle.Width(14).Render("BALANCE"),
	)
	sep := cellDimStyle.Render(strings.Repeat("·", 58))

	rows := []string{header, sep}
	for _, a := range accounts {
		row := lipgloss.JoinHorizontal(lipgloss.Left,
			cellDimStyle.Width(44).Render(a.Name),
			cellStyle.Width(14).Render(utils.FormatEUR(a.Value)),
		)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, section, strings.Join(rows, "\n"))
}

// ── Dashboard footer ──────────────────────────────────────────────────────────

func renderFooter() string {
	keys := []struct{ key, desc string }{
		{"r", "refresh"},
		{"w", "watchlist"},
		{"y", "Yahoo Finance"},
		{"q", "quit"},
	}

	parts := make([]string, 0, len(keys)*3)
	for i, k := range keys {
		parts = append(parts, keyStyle.Render(k.key))
		parts = append(parts, footerStyle.Render(" "+k.desc))
		if i < len(keys)-1 {
			parts = append(parts, footerStyle.Render("  ·  "))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

// ── Watchlist view ────────────────────────────────────────────────────────────

func renderWatchlist(sections []fetch.WatchSection, width int) string {
	now := time.Now().Format("01/02/2006  15:04:05")
	logo := logoStyle.Render("▸ WATCHLIST")
	ts := subtitleStyle.Render(now)

	gap := width - lipgloss.Width(logo) - lipgloss.Width(ts) - 6
	if gap < 1 {
		gap = 1
	}
	top := lipgloss.JoinHorizontal(lipgloss.Top, logo, strings.Repeat(" ", gap), ts)
	headerLine := headerLineStyle.Render(strings.Repeat("─", width-6))
	header := lipgloss.JoinVertical(lipgloss.Left, top, headerLine)

	// Column widths: name(24) symbol(10) price(12) 1d(9) 7d(9) 30d(9) 90d(9) 365d(9) ath(12) athdiff(10)
	colHeader := lipgloss.JoinHorizontal(lipgloss.Left,
		tableHeaderStyle.Width(24).Render("NAME"),
		tableHeaderStyle.Width(10).Render("SYMBOL"),
		tableHeaderStyle.Width(12).Render("PRICE"),
		tableHeaderStyle.Width(9).Render("1D"),
		tableHeaderStyle.Width(9).Render("7D"),
		tableHeaderStyle.Width(9).Render("30D"),
		tableHeaderStyle.Width(9).Render("90D"),
		tableHeaderStyle.Width(9).Render("1Y"),
		tableHeaderStyle.Width(14).Render("ATH (1Y)"),
		tableHeaderStyle.Width(10).Render("vs ATH"),
	)

	lines := []string{header, ""}

	for _, sec := range sections {
		secTitle := sectionStyle.Render(strings.ToUpper(sec.Name))
		secLine := sectionLineStyle.Render(strings.Repeat("─", 60))
		lines = append(lines, secTitle, secLine, colHeader)

		sep := cellDimStyle.Render(strings.Repeat("·", 115))
		lines = append(lines, sep)

		for _, p := range sec.Items {
			lines = append(lines, renderWatchRow(p))
		}
		lines = append(lines, "")
	}

	lines = append(lines, renderWatchlistFooter())

	return strings.Join(lines, "\n")
}

func renderWatchRow(p fetch.WatchPerf) string {
	perfCell := func(v float64, w int) string {
		s := utils.FormatPercent(v)
		if v > 0 {
			return posStyle.Width(w).Render(s)
		}
		return negStyle.Width(w).Render(s)
	}

	athDiffCell := func(v float64) string {
		if v == 0 {
			return posStyle.Width(10).Render("ATH")
		}
		return negStyle.Width(10).Render(fmt.Sprintf("%.2f%%", v))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		cellStyle.Width(24).Render(utils.Truncate(p.Name, 23)),
		symbolStyle.Width(10).Render(p.Symbol),
		cellStyle.Width(12).Render(utils.FormatEUR(p.Price)),
		perfCell(p.Perf1d, 9),
		perfCell(p.Perf7d, 9),
		perfCell(p.Perf30d, 9),
		perfCell(p.Perf90d, 9),
		perfCell(p.Perf365d, 9),
		cellDimStyle.Width(14).Render(utils.FormatEUR(p.ATH)),
		athDiffCell(p.ATHDiff),
	)
}

func renderWatchlistFooter() string {
	keys := []struct{ key, desc string }{
		{"r", "refresh"},
		{"b / esc", "back"},
		{"q", "quit"},
	}

	parts := make([]string, 0, len(keys)*3)
	for i, k := range keys {
		parts = append(parts, keyStyle.Render(k.key))
		parts = append(parts, footerStyle.Render(" "+k.desc))
		if i < len(keys)-1 {
			parts = append(parts, footerStyle.Render("  ·  "))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

// ── Loading / Error ───────────────────────────────────────────────────────────

func renderLoading() string {
	return loadingStyle.Render("⠿  Fetching quotes…")
}

func renderError(err error) string {
	return errorStyle.Render(fmt.Sprintf("✗  Error: %v\n\nPress r to retry.", err))
}
