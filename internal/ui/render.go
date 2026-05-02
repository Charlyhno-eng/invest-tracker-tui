package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"invest-tracker-tui/internal/config"
	"invest-tracker-tui/internal/fetch"
	"invest-tracker-tui/internal/utils"
)


func renderKPI(title, value string, width int) string {
	return cardStyle.Width(width).Render(
		kpiTitleStyle.Render(title) + "\n" +
			kpiValueStyle.Render(value),
	)
}

func renderQuotesTable(quotes []fetch.Quote) string {
	var lines []string

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerStyle.Width(40).Render("Actif"),
		headerStyle.Width(12).Render("Prix"),
		headerStyle.Width(10).Render("Var %"),
		headerStyle.Width(12).Render("Parts"),
		headerStyle.Width(14).Render("Valeur"),
	)
	lines = append(lines, header)

	for _, q := range quotes {
		change := utils.FormatPercent(q.ChangePercent)
		if q.ChangePercent >= 0 {
			change = posStyle.Width(10).Render(change)
		} else {
			change = negStyle.Width(10).Render(change)
		}

		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(40).Render(utils.Truncate(q.Name, 40)),
			lipgloss.NewStyle().Width(12).Render(utils.FormatEUR(q.Price)),
			change,
			lipgloss.NewStyle().Width(12).Render(utils.FormatShares(q.Shares)),
			lipgloss.NewStyle().Width(14).Render(utils.FormatEUR(q.Value)),
		)
		lines = append(lines, line)
	}

	return cardStyle.Render(strings.Join(lines, "\n"))
}

func renderAccounts(accounts []config.Account) string {
	var lines []string
	lines = append(lines, headerStyle.Render("Comptes"))

	for _, a := range accounts {
		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(40).Render(a.Name),
			lipgloss.NewStyle().Width(14).Render(utils.FormatEUR(a.Value)),
		)
		lines = append(lines, line)
	}

	return cardStyle.Render(strings.Join(lines, "\n"))
}
