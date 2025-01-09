package ui

import "github.com/charmbracelet/lipgloss"

const (
	CheckPassIcon  = "✅"
	CheckFailIcon  = "❌"
	CheckErrorIcon = "🟠"
	TreeVertical   = "│"
	TreeBranch     = "├──"
	TreeLeaf       = "└──"
)

// Styles contains all the styles used in the UI
type Styles struct {
	Success     lipgloss.Style
	Error       lipgloss.Style
	Warning     lipgloss.Style
	OutputBox   lipgloss.Style
	ErrorBox    lipgloss.Style
	GroupHeader lipgloss.Style
	TreeBranch  lipgloss.Style
}

// NewStyles creates a new Styles instance
func NewStyles() *Styles {
	return &Styles{
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")),

		OutputBox: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1).
			MarginLeft(4),

		ErrorBox: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("9")).
			Padding(0, 1).
			MarginLeft(4),

		GroupHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginTop(1),

		TreeBranch: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
	}
}
