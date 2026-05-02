package ui

import "github.com/charmbracelet/lipgloss"

var (
	bgStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			Background(lipgloss.Color("#F5F3FF")).
			Padding(0, 1).
			MarginBottom(1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#D4D4D8")).
			Padding(1, 2)

	kpiTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	kpiValueStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#71717A"))
	posStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#15803D")).Bold(true)
	negStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#BE123C")).Bold(true)
)
