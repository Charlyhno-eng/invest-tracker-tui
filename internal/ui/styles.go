package ui

import "github.com/charmbracelet/lipgloss"

// Color palette — dark financial terminal theme.
const (
	colorCyan      = "#00D4FF"
	colorCyanDim   = "#007A99"
	colorAmber     = "#FFB800"
	colorAmberDim  = "#7A5800"
	colorGreen     = "#00E676"
	colorRed       = "#FF3D57"
	colorTextHi    = "#E8F4FD"
	colorTextMid   = "#7A9BB5"
	colorTextLow   = "#3A5870"
	colorBorder    = "#1E2D40"
	colorBorderHi  = "#2A4A6B"
)

var (
	bgStyle = lipgloss.NewStyle().Padding(1, 3)

	// Header styles.
	logoStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorCyan))
	subtitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorTextMid))
	headerLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorBorder))

	// KPI card styles.
	kpiCardStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(0, 2)

	kpiCardHiStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorCyan)).
			Padding(0, 2)

	kpiLabelStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(colorTextMid))
	kpiValueStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorTextHi))
	kpiTotalValueStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorCyan))

	// Section label styles.
	sectionStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorAmber))
	sectionLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAmberDim))

	// Table styles.
	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorTextMid))
	cellStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(colorTextHi))
	cellDimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorTextMid))
	symbolStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyanDim))

	// Asset type badge styles.
	badgeStockStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAmber))
	badgeCryptoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan))

	// Change column styles.
	posStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorGreen))
	negStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorRed))

	// Footer styles.
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorTextLow))
	keyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan))

	// State styles.
	loadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan)).Padding(2, 4)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorRed)).Padding(2, 4)
)
