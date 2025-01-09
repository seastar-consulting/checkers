package ui

import "github.com/charmbracelet/lipgloss"

const (
	CheckPassIcon  = "✅"
	CheckFailIcon  = "❌"
	CheckErrorIcon = "🟠"
)

// Styles holds all the UI styles used in the application
type Styles struct {
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	ErrorBox   lipgloss.Style
	OutputBox  lipgloss.Style
}

// NewStyles creates a new Styles instance with default styling
func NewStyles() *Styles {
	return &Styles{
		Success: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")), // Green

		Error: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9")), // Red

		Warning: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("208")), // Orange

		ErrorBox: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("9")). // Red border
			Background(lipgloss.Color("1")).       // Dark red background
			Foreground(lipgloss.Color("15")).      // White text
			BorderStyle(lipgloss.RoundedBorder()).
			MarginLeft(4),

		OutputBox: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			MarginLeft(4),
	}
}
