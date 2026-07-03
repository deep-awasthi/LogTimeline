package ui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F6F6F6")).
			Background(lipgloss.Color("#385F71")).
			Padding(0, 1)
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4E6E81")).
			Padding(1, 2)
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EAEAEA")).
			Background(lipgloss.Color("#5E6973")).
			Padding(0, 1)
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8A9299"))
	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDE3E7"))
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#102027")).
				Background(lipgloss.Color("#B7E4C7"))
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#B7E4C7"))
	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9AA6B2"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB4A2"))
)
