package styles

import "github.com/charmbracelet/lipgloss"

var (
	DescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Padding(0, 1)
	TitleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	GreenStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	RedStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	SelectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	ItemStyle         = lipgloss.NewStyle()
	DocStyle          = lipgloss.NewStyle().Margin(1, 2)
	FadeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	FocusedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	BlurredStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Italic(true).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4F6892"))
	BoxStyle          = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#89b4fa")).
				Margin(0, 0)
)
